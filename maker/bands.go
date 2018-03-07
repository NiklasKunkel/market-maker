package maker

import(
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"github.com/sirupsen/logrus"
)

//Globals
var validCombos = [][]*Order{}
//var allBands = make(AllBands)

///////////////////////////////////
//         BANDS
///////////////////////////////////
type AllBands map[string]Bands
type Bands struct {
	BuyBands 	[]BuyBand 	`json:"buyBands"`
	SellBands 	[]SellBand 	`json:"sellBands"`
}

//Load bands from bands.json file
func (allBands AllBands) LoadBands() (bool) {
	//clear existing bands
	for _, bands := range allBands {
		bands.BuyBands = nil
		bands.SellBands = nil
	}
	//lookup $GOPATH
	goPath, ok := os.LookupEnv("GOPATH")
	if ok != true {
		log.WithFields(logrus.Fields{"function": "LoadBands"}).Fatal("$GOPATH Env Variable not set")
	}
	//construct path to bands.json
	bandPath := goPath + "/src/github.com/niklaskunkel/market-maker/bands.json"
	//read bands.json
	raw, err := ioutil.ReadFile(bandPath)
	if err != nil {
		log.WithFields(logrus.Fields{"function": "LoadBands", "error": err.Error()}).Error("Unable to read bandsNew.json")
		return false
	}
	//load json into memory
	err = json.Unmarshal(raw, &allBands)
	if err != nil {
		log.WithFields(logrus.Fields{"function": "LoadBands", "error": err.Error()}).Error("Loading bandsNew failed during Unmarshal")
		return false
	}
	//print bands
	allBands.PrintAllBands()

	//verify bands
	for _, bands := range allBands {
		if(!bands.VerifyBands()) {
			return false
		}
	}
	return true
}

func (allBands AllBands) PrintAllBands() {
	for tokenPair, bands := range allBands {
		log.Debug(tokenPair + "Bands:")
		bands.PrintBands()
	}
}

//Print all bands of a specific token pair
func (bands *Bands) PrintBands() {
	log.Debug("Buy Bands:")
	for i, bBand := range bands.BuyBands {
		bBand.PrintBand(i)
	}
	log.Debug("Sell Bands:")
	for i, sBand := range bands.SellBands {
		sBand.PrintBand(i)
	}
}

//Verify band parameters
func (bands *Bands) VerifyBands() (bool) {
	for  _, bBand := range bands.BuyBands {
		err := bBand.VerifyBand()
		if err != nil {
			log.WithFields(logrus.Fields{"function": "VerifyBands", "band": bBand, "error": err.Error()}).Error("Buy band verification failed")
			return false
		}
	}
	for _, sBand := range bands.SellBands {
		err := sBand.VerifyBand()
		if err != nil {
			log.WithFields(logrus.Fields{"function": "VerifyBands", "band": sBand, "error": err.Error()}).Error("Sell band verification failed")
			return false
		}
	}
	if(bands.BandsOverlap()) {
		log.WithFields(logrus.Fields{"function": "VerifyBands"}).Error("Band verification failed due to overlapping bands")
		return false
	}
	return true
}

func (bands *Bands) BandsOverlap() (bool) {
	for _, band1 := range bands.BuyBands {
		for _, band2 := range bands.BuyBands {
			if band1 == band2 {
				continue
			}
			if (band1.MinMargin < band2.MaxMargin && band2.MinMargin < band1.MaxMargin) {
				return true
			}
		}
	}
	for _, band1 := range bands.SellBands {
		for _, band2 := range bands.SellBands {
			if band1 == band2 {
				continue
			}
			if (band1.MinMargin < band2.MaxMargin && band2.MinMargin < band1.MaxMargin) {
				return true
			}
		}
	}
	return false
}

//Returns buy orders which need to be cancelled to bring total amount within all buy bands below maximum
func (bands *Bands) ExcessiveBuyOrders(buyOrders []*Order, refPrice float64) (cancellableBuyOrders []*Order){
	for _, buyBand := range bands.BuyBands {
		for _, order := range buyBand.ExcessiveOrders(buyOrders, refPrice) {
			cancellableBuyOrders = append(cancellableBuyOrders, order)
		}
	}
	return cancellableBuyOrders
}

//Return sell orders which need to be cancelled to bring total amount within all sell bands below maximum
func (bands *Bands) ExcessiveSellOrders(sellOrders []*Order, refPrice float64) (cancellableSellOrders []*Order) {
	for _, sellBand := range bands.SellBands {
		for _, order := range sellBand.ExcessiveOrders(sellOrders, refPrice) {
			cancellableSellOrders = append(cancellableSellOrders, order)
		}
	}
	return cancellableSellOrders
}

//Returns orders which do not fall into any buy or sell band
func (bands *Bands) OutsideOrders(buyOrders []*Order, sellOrders []*Order, refPrice float64) (outsideOrders []*Order) {
	for _, buyOrder := range buyOrders {
		inBand := false
		for _, band := range bands.BuyBands {
			if (band.Includes(buyOrder.Price, refPrice)) {
				inBand = true
			}
		}
		if (!inBand) {
			outsideOrders = append(outsideOrders, buyOrder)
		}
	}
	for _, sellOrder := range sellOrders {
		inBand := false
		for _, band := range bands.SellBands {
			if (band.Includes(sellOrder.Price, refPrice)) {
				inBand = true
			}
		}
		if (!inBand) {
			outsideOrders = append(outsideOrders, sellOrder)
		}
	}
	return outsideOrders
}

func (bands Bands) CancellableOrders(buyOrders []*Order, sellOrders []*Order, refPrice float64) (ordersToCancel []*Order) {
	ordersToCancel = append(ordersToCancel, bands.ExcessiveBuyOrders(buyOrders, refPrice)...)
	ordersToCancel = append(ordersToCancel, bands.ExcessiveSellOrders(sellOrders, refPrice)...)
	ordersToCancel = append(ordersToCancel, bands.OutsideOrders(buyOrders, sellOrders, refPrice)...)
	return ordersToCancel
}

///////////////////////////////////
//         BAND
///////////////////////////////////
type Band struct {
	MinMargin 	float64 	`json:"minMargin"`
	AvgMargin 	float64 	`json:"avgMargin"`
	MaxMargin 	float64 	`json:"maxMargin"`
	MinAmount 	float64 	`json:"minAmount"`
	AvgAmount 	float64 	`json:"avgAmount"`
	MaxAmount 	float64 	`json:"maxAmount"`
	DustCutoff 	float64 	`json:"dustCutoff"`
}

type BandType interface {
	Includes(float64, float64) bool
}

func (band *Band) VerifyBand() (error) {
	if (band.MinMargin <= float64(0) || band.MinMargin >= float64(1) || band.MinMargin > band.AvgMargin) {
		return fmt.Errorf("Error: Band verification failed, MinMargin(%f) > AvgMargin(%f) and must not equal zero.\n", band.MinMargin, band.AvgMargin)
	}
	if (band.AvgMargin <= float64(0) || band.AvgMargin >= float64(1) || band.AvgMargin > band.MaxMargin) {
		return fmt.Errorf("Error: Band verification failed, AvgMargin(%f) > MaxMargin(%f) and must not equal zero.\n", band.AvgMargin, band.MaxMargin)
	}
	if (band.MaxMargin  <= float64(0) || band.MaxMargin >= float64(1) || band.MinMargin >= band.MaxMargin) {
		return fmt.Errorf("Error: Band verification failed, MinMargin(%f) >= MaxMargin(%f) and must not equal zero.\n", band.MinMargin, band.MaxMargin)
	}
	if (band.MinAmount <= float64(0) || band.MinAmount > band.AvgAmount) {
		return fmt.Errorf("Error: Band verification failed, MinAmount(%f) > AvgAmount(%f) and must not equal zero.\n", band.MinAmount, band.AvgAmount)
	}
	if (band.AvgAmount <= float64(0) || band.AvgAmount > band.MaxAmount) {
		return fmt.Errorf("Error: Band verification failed, AvgAmount(%f) > MaxAmount(%f) and must not equal zero.\n", band.AvgAmount, band.MaxAmount)
	}
	if (band.MaxAmount <= float64(0) || band.MinAmount > band.MaxAmount) {
		return fmt.Errorf("Error: Band verification failed, MinAmount(%f) > MaxAmount(%f) and must not equal zero.\n", band.MinAmount, band.MaxAmount)
	}
	return nil
}

//Returns orders which need to be cancelled to bring the total
//order amount in the band below the maximum
func (band *Band) ExcessiveOrders(orders []*Order, refPrice float64, bandType BandType) ([]*Order) {
	ordersInBand := []*Order{}
	for _, order := range orders {
		included := false
		if t, ok := bandType.(BandType); ok {
			included = t.Includes(order.Price, refPrice)
		} else {
			included = band.Includes(order.Price, refPrice)
		}
		if (included) {
			ordersInBand = append(ordersInBand, order)
		}
	}
	for _, orderInBand := range ordersInBand {
		log.WithFields(logrus.Fields{"function": "ExcessiveOrders", "refPrice": refPrice, "bandType": bandType, "orderId": orderInBand.OrderId, "RemQuantity": orderInBand.RemQuantity,}).Debug("Order Included in Band")
	}
	if (band.TotalAmount(ordersInBand) > band.MaxAmount) {
		log.WithFields(logrus.Fields{"function": "ExcessiveOrders", "refPrice": refPrice, "bandType": bandType, "totalAmount": band.TotalAmount(ordersInBand), "maxAmount": band.MaxAmount}).Info("Total Order Amount Exceeded, finding orders to cancel...")
		//log.WithFields(logrus.Fields{}).Debug("All Combinations")
		for size, _ := range ordersInBand {
			band.GetAllCombinationsOfSizeN(ordersInBand, size + 1)
		}
		log.WithFields(logrus.Fields{"function": "ExcessiveOrders", "refPrice": refPrice, "bandType": bandType}).Debug("Valid Combinations of Orders:")
		for comboNum, combo := range validCombos {
			for _, element := range combo {
				log.WithFields(logrus.Fields{"function": "ExcessiveOrders", "refPrice": refPrice, "bandType": bandType, "comboNumber": comboNum, "comboSize": len(combo),"orderId": element.OrderId, "price": element.Price, "remQuantity": element.RemQuantity}).Debug("order in valid combo")
			}
		}
		//filter combos by largest length - this means we have to cancel less orders (saves gas for dex)
		//if one or more combos share the largest length choose the one with the higher total amount
		maxLength := 0
		maxIndex := -1
		for i, combo := range validCombos {
			comboLength := len(combo)
			if (comboLength == maxLength) {
				if (band.TotalAmount(combo) > band.TotalAmount(validCombos[maxIndex])) {
					maxIndex = i
				}
			} else if (comboLength > maxLength) {
				maxIndex = i
				maxLength = comboLength
			}
		}
		log.WithFields(logrus.Fields{"function": "ExcessiveOrders", "numberOfOrdersInCombo": maxLength, "comboNumber": maxIndex}).Debug("Best combination of orders found")
		//Kill all orders not in our best combination
		ordersToKill := []*Order{}
		for _, order := range ordersInBand {
			keepOrder := false
			for _, orderToKeep := range validCombos[maxIndex] {
				if (order == orderToKeep) {
					keepOrder = true
				}
			}
			if (keepOrder == false) {
				ordersToKill = append(ordersToKill, order)
			}
		}
		//clear validCombos global
		validCombos = nil
		for _, killOrder := range ordersToKill {
			log.WithFields(logrus.Fields{"function": "ExcessiveOrders", "orderId": killOrder.OrderId, "price": killOrder.Price, "remQuantity": killOrder.RemQuantity}).Debug("Order flagged for cancellation")
		}
		return ordersToKill
	} else {
		return []*Order{}
	}
}

func (band *Band) GetAllCombinationsOfSizeN(input []*Order, comboSize int) {
	length := len(input)
	output := make([]*Order, comboSize)
	band.CombinationUtil(input, output, 0, length - 1, 0, comboSize)
}

func (band *Band) CombinationUtil(input []*Order, output []*Order, start int, end int, index int, comboSize int) {
	//Print Combo
	if (index == comboSize) {
		total := band.TotalAmount(output)
		if (total >= band.MinAmount && total < band.MaxAmount) {
			temp := make([]*Order, comboSize)
			copy(temp, output)
			validCombos = append(validCombos, temp)
		}
		/*for i, element := range output {
			log.WithFields(logrus.Fields{"function": "CombinationUtil", "comboSize": comboSize, "comboNumber": i + 1, "orderId": element.OrderId, "price": element.Price, "remQuantity": element.RemQuantity}).Debug("Order in possible combo")
		}*/
		return
	}

	//Replace output[index] with all possible elements of input
    for i := start; i <= end && ((end - i + 1) >= (comboSize - index)); i++ {
    	output[index] = input[i]
    	band.CombinationUtil(input, output, i + 1, end, index + 1, comboSize)
    }
}

func (band *Band) Includes(orderPrice float64, refPrice float64) (bool) {
	//raise virtual method exception
	log.WithFields(logrus.Fields{"function": "Includes", "band": band}).Fatal("Using base class Includes(), this should never happen!")
	return true
}

//Returns the total amount of all the orders
func (band *Band) TotalAmount(orders []*Order) (total float64) {
	for _, order := range orders {
		total += order.RemQuantity
	}
	return total
}

func (band *Band) PrintBand(i int) {
	log.WithFields(logrus.Fields{"BandNum": i + 1, "minMargin": band.MinMargin, "avgMargin": band.AvgMargin, "maxMargin": band.MaxMargin, "minAmount": band.MinMargin, "avgAmount": band.AvgAmount, "maxAmount": band.MaxAmount, "dustCutoff": band.DustCutoff}).Debug()
}

///////////////////////////////////
//         BUY BAND
///////////////////////////////////
type BuyBand struct {
	Band
}

func (band *BuyBand) Includes(orderPrice float64, refPrice float64) (bool) {
	log.WithFields(logrus.Fields{"function": "Includes", "band": band}).Debug("Using Includes() from buy band")
	minPrice := band.ApplyMargin(refPrice, band.MinMargin)
	maxPrice := band.ApplyMargin(refPrice, band.MaxMargin)
	isIncluded := (orderPrice >= maxPrice) && (orderPrice <= minPrice)
	log.WithFields(logrus.Fields{"function": "Includes", "bandType": "buy band", "minMargin": band.MinMargin, "maxMargin": band.MaxMargin, "minPrice": minPrice, "maxPrice": maxPrice, "orderPrice": orderPrice, "included": isIncluded}).Debug("Checking if order is in buy band...")
	return isIncluded
}

func (band *BuyBand) AvgPrice(refPrice float64) (float64) {
	return band.ApplyMargin(refPrice, band.AvgMargin)
}

func (band *BuyBand) ApplyMargin(price float64, margin float64) (float64) {
	return price * (1 - margin)
}

func (band *BuyBand) ExcessiveOrders(orders []*Order, refPrice float64) ([]*Order) {
	return band.Band.ExcessiveOrders(orders, refPrice, band)
}

///////////////////////////////////
//         SELL BAND
///////////////////////////////////
type SellBand struct {
	Band
}

func (band *SellBand) Includes(orderPrice float64, refPrice float64) (bool) {
	log.WithFields(logrus.Fields{"function": "Includes", "band": band}).Debug("Using Includes() from sell band")
	minPrice := band.ApplyMargin(refPrice, band.MinMargin)
	maxPrice := band.ApplyMargin(refPrice, band.MaxMargin)
	isIncluded := (orderPrice >= minPrice) && (orderPrice <= maxPrice)
	log.WithFields(logrus.Fields{"function": "Includes", "bandType": "sell band", "minMargin": band.MinMargin, "maxMargin": band.MaxMargin, "minPrice": minPrice, "maxPrice": maxPrice, "orderPrice": orderPrice, "included": isIncluded}).Debug("Checking if order is in sell band...")
	return isIncluded
}

func (band *SellBand) AvgPrice(refPrice float64) (float64) {
	return band.ApplyMargin(refPrice, band.AvgMargin)
}

func (band *SellBand) ApplyMargin(price float64, margin float64) (float64) {
	return price * (1 + margin)
}

func (band *SellBand) ExcessiveOrders(orders []*Order, refPrice float64) ([]*Order) {
	return band.Band.ExcessiveOrders(orders, refPrice, band)
}