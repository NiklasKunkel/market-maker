package maker

import(
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

var validCombos = [][]*Order{}

///////////////////////////////////
//         BANDS
///////////////////////////////////
type Bands struct {
	BuyBands 	[]BuyBand 	`json:"buyBands"`
	SellBands 	[]SellBand 	`json:"sellBands"`
}

type Order struct {
	Code 			string 	
	OrderId 		string
	Side 			int64
	Price 			float64
	InitQuantity 	float64
	RemQuantity 	float64
	Status 			int64
	StatusDesc 		string
	TxSeqNo 		int64
	Type 			int64
	Date 			string
}

//Load bands from bands.json file
func (bands *Bands) LoadBands() (error) {
	absPath, _ := filepath.Abs("/Users/nkunkel/Programming/Go/src/github.com/niklaskunkel/market-maker/bands.json")
	raw, err := ioutil.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("Error: Band loading failed, %s\n", err.Error())
	}
	err = json.Unmarshal(raw, bands)
	if err != nil {
		return fmt.Errorf("Error: Band loading failed, %s\n", err.Error())
	}
	for _, band := range bands.BuyBands  {
		err = band.VerifyBand()
		if err != nil {
			return fmt.Errorf("%s\n", err.Error())
		}
	}
	for _, band :=  range bands.SellBands {
		err = band.VerifyBand()
		if err != nil {
			return fmt.Errorf("%s\n", err.Error())
		}
	}
	return nil
}

//Print all bands
func (bands *Bands) PrintBands() {
	fmt.Printf("Buy Bands:\n")
	for i, bBand := range bands.BuyBands {
		fmt.Printf("#%d\n", i + 1)
		bBand.PrintBand()
	}
	fmt.Printf("Sell Bands:\n")
	for i, sBand := range bands.SellBands {
		fmt.Printf("#%d\n", i + 1)
		sBand.PrintBand()
	}
}

//Verify band parameters
func (bands *Bands) VerifyBands() (error) {
	for  _, bBand := range bands.BuyBands {
		err := bBand.VerifyBand()
		if err != nil {
			return err
		}
	}
	for _, sBand := range bands.SellBands {
		err := sBand.VerifyBand()
		if err != nil {
			return err
		}
	}
	if(bands.BandsOverlap()) {
		return fmt.Errorf("Error during band verification: overlapping bands\n")
	}
	return nil
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
func (bands *Bands) ExcessiveBuyOrders(buyOrders []*Order, targetPrice float64) (killBuyOrders []*Order){
	for _, buyBand := range bands.BuyBands {
		for _, order := range buyBand.ExcessiveOrders(buyOrders, targetPrice) {
			killBuyOrders = append(killBuyOrders, order)
		}
	}
	return killBuyOrders
}

//Return sell orders which need to be cancelled to bring total amount within all sell bands below maximum
func (bands *Bands) ExcessiveSellOrders(sellOrders []*Order, targetPrice float64) (killSellOrders []*Order) {
	for _, sellBand := range bands.SellBands {
		for _, order := range sellBand.ExcessiveOrders(sellOrders, targetPrice) {
			killSellOrders = append(killSellOrders, order)
		}
	}
	return killSellOrders
}

//Returns orders which do not fall into any buy or sell band
func (bands *Bands) OutsideOrders(orders []*Order, targetPrice float64) (outsideOrders []*Order) {
	for _, order := range orders {
		if order.Side == 0 {
			inBand := false
			for _, band := range bands.BuyBands {
				if (band.Includes(order.Price, targetPrice)) {
					inBand = true
				}
			}
			if (!inBand) {
				outsideOrders = append(outsideOrders, order)
			}
		} else if order.Side == 1 {
			inBand := false
			for _, band := range bands.SellBands {
				if (band.Includes(order.Price, targetPrice)) {
					inBand = true
				}
			}
			if (!inBand) {
				outsideOrders = append(outsideOrders, order)
			}
		}
	}
	return outsideOrders
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
func (band *Band) ExcessiveOrders(orders []*Order, targetPrice float64) ([]*Order) {
	ordersInBand := []*Order{}
	for _, order := range orders {
		if (band.Includes(order.Price, targetPrice)) {
			ordersInBand = append(ordersInBand, order)
		}
	}
	fmt.Printf("Orders Included:\n")
	for _, orderInBand := range ordersInBand {
		fmt.Printf("%f\n", orderInBand.RemQuantity)
	}
	if (band.TotalAmount(ordersInBand) > band.MaxAmount) {
		fmt.Printf("All Combinations:\n")
		for size, _ := range orders {
			band.GetAllCombinationsOfSizeN(orders, size + 1)
		}
		fmt.Printf("\nValid Combinations:\n")
		for _, combo := range validCombos {
			for _, element := range combo {
				fmt.Printf("%f ", element.RemQuantity)
			}
			fmt.Printf("\n")
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

		//Debug - delete later
		fmt.Printf("Max Length = %d\n", maxLength)
		fmt.Printf("Max Index = %d\n", maxIndex)

		ordersToKill := []*Order{}

		for _, order := range orders {
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

		//Debug - delete later
		for _, killOrder := range ordersToKill {
			fmt.Printf("Kill Order: %+v\n", killOrder)
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
		//You can put all your logic for checking combos here, don't even bother storing them.
		total := band.TotalAmount(output)
		if (total >= band.MinAmount && total < band.MaxAmount) {
			temp := make([]*Order, comboSize)
			copy(temp, output)
			fmt.Printf("Appending...")
			validCombos = append(validCombos, temp)
		}

		for _, element := range output {
			fmt.Printf("%f ", element.RemQuantity)
		}
		fmt.Printf("\n")
		return
	}

	//Replace output[index] with all possible elements of input
    for i := start; i <= end && ((end - i + 1) >= (comboSize - index)); i++ {
    	output[index] = input[i]
    	band.CombinationUtil(input, output, i + 1, end, index + 1, comboSize)
    }
}

func (band *Band) Includes(orderPrice float64, targetPrice float64) (bool) {
	//raise virtual method exception
	fmt.Printf("Using Includes() from base class\n")
	return true
}

//Returns the total amount of all the orders
//Warning - slice of orders passed must be exclusive bids or exclusively asks.
func (band *Band) TotalAmount(orders []*Order) (total float64) {
	for _, order := range orders {
		total += order.RemQuantity
	}
	return total
}

func (band *Band) PrintBand() {
	fmt.Printf("MinMargin = %f\n", band.MinMargin)
	fmt.Printf("AvgMargin = %f\n", band.AvgMargin)
	fmt.Printf("MaxMargin = %f\n", band.MaxMargin)
	fmt.Printf("MinAmount = %f\n", band.MinAmount)
	fmt.Printf("AvgAmount = %f\n", band.AvgAmount)
	fmt.Printf("MaxAmount = %f\n", band.MaxAmount)
	fmt.Printf("DustCutoff = %f\n\n", band.DustCutoff)
}

///////////////////////////////////
//         BUY BAND
///////////////////////////////////
type BuyBand struct {
	Band
}

func (band *BuyBand) Includes(orderPrice float64, targetPrice float64) (bool) {
	fmt.Printf("Using Includes() from BuyBand class\n")
	fmt.Printf("BuyBand: MinMargin = %f, MaxMargin = %f, OrderPrice = %f\n", band.MinMargin, band.MaxMargin, orderPrice )
	minPrice := band.ApplyMargin(targetPrice, band.MinMargin)
	maxPrice := band.ApplyMargin(targetPrice, band.MaxMargin)
	fmt.Printf("BuyBand: MinPrice = %f, MaxPrice = %f, OrderPrice = %f\n", minPrice, maxPrice, orderPrice)
	return (orderPrice >= maxPrice) && (orderPrice <= minPrice)
}

func (band *BuyBand) AvgPrice(targetPrice float64) (float64) {
	return band.ApplyMargin(targetPrice, band.AvgMargin)
}

func (band *BuyBand) ApplyMargin(price float64, margin float64) (float64) {
	return price * (1 - margin)
}

///////////////////////////////////
//         SELL BAND
///////////////////////////////////
type SellBand struct {
	Band
}

func (band *SellBand) Includes(orderPrice float64, targetPrice float64) (bool) {
	fmt.Printf("Using Includes() from SellBand class\n")
	fmt.Printf("SellBand: MinMargin = %f, MaxMargin = %f, OrderPrice = %f\n", band.MinMargin, band.MaxMargin, orderPrice)
	minPrice := band.ApplyMargin(targetPrice, band.MinMargin)
	maxPrice := band.ApplyMargin(targetPrice, band.MaxMargin)
	fmt.Printf("SellBand: MinPrice = %f, MaxPrice = %f, OrderPrice = %f\n", minPrice, maxPrice, orderPrice)
	return (orderPrice >= minPrice) && (orderPrice <= maxPrice)
}


func (band *SellBand) AvgPrice(targetPrice float64) (float64) {
	return band.ApplyMargin(targetPrice, band.AvgMargin)
}

func (band *SellBand) ApplyMargin(price float64, margin float64) (float64) {
	return price * (1 + margin)
}