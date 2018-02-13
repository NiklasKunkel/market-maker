package maker

import	(
	"testing"
	"github.com/stretchr/testify/assert"
)

//BANDS
func Test_Bands_LoadBands(t *testing.T) {
	bands := new(Bands)					//create bands instance
	assert.True(t, bands.LoadBands())	//load bands from bands.json
	assert.NotEmpty(t, bands.BuyBands)	//check buy bands exist
	assert.NotEmpty(t, bands.SellBands)	//check sell bands exist
}

func Test_Bands_VerifyBands(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	assert.True(t, bands.VerifyBands())	//verify if bands have correct paramters
}

func Test_Bands_BandsOverlap1(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	assert.False(t, bands.BandsOverlap())	//check if bands overlap - they should not
}

func Test_Bands_BandsOverlap2(t* testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	bands.BuyBands = append(bands.BuyBands, bands.BuyBands[0])	//clone buy band
	bands.BuyBands[1].MinMargin = .005							//modify band to fall within range of band[0]
	assert.True(t, bands.BandsOverlap())						//check if bands overlap - they should
}

func Test_Bands_ExcessiveBuyOrders(t *testing.T) {

}

func Test_Bands_ExcessiveSellOrders(t *testing.T) {

}

//Test if bid on boundary of minMargin is in-band
func Test_Bands_OutsideOrders1(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.BuyBands = []BuyBand{BuyBand{Band{0.002344, 0.004689, 0.009378, 10.0, 40.0, 80.0, 0.0}}}
	buyOrders := []*Order{&Order{"DAIUSD", "BK01", 0, 0.997656, 50.0, 20.0, 1, "New", 0, 0, "1515755942"}}	//create in-band bid order
	sellOrders := []*Order{}
	refPrice := 1.00					//set ref price of asset to 1
	assert.Empty(t, bands.OutsideOrders(buyOrders, sellOrders, refPrice))	//assert order is within boundary
}

//Test if bid on boundary of maxMargin is in-band
func Test_Bands_OutsideOrders2(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.BuyBands = []BuyBand{BuyBand{Band{0.002344, 0.004689, 0.009378, 10.0, 40.0, 80.0, 0.0}}}
	buyOrders := []*Order{&Order{"DAIUSD", "BK01", 0, 0.990622, 50.0, 20.0, 1, "New", 0, 0, "1515755942"}}	//create in-band bid order
	sellOrders := []*Order{}
	refPrice := 1.00					//set ref price of asset to 1
	assert.Empty(t, bands.OutsideOrders(buyOrders, sellOrders, refPrice))	//assert order is within boundary
}

//Test if bid on minMargin++ is in-band
func Test_Bands_OutsideOrders3(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.BuyBands = []BuyBand{BuyBand{Band{0.002344, 0.004689, 0.009378, 10.0, 40.0, 80.0, 0.0}}}
	buyOrders := []*Order{&Order{"DAIUSD", "BK01", 0, 0.997657, 50.0, 20.0, 1, "New", 0, 0, "1515755942"}}	//create outside-band bid order
	sellOrders := []*Order{}
	refPrice := 1.00					//set ref price of asset to 1
	assert.NotEmpty(t, bands.OutsideOrders(buyOrders, sellOrders, refPrice))				//assert order is out of bondary
	assert.Contains(t, bands.OutsideOrders(buyOrders, sellOrders, refPrice), buyOrders[0])	//assert order did not get corrupted
}

//Test if bid on maxMargin-- is in-band
func Test_Bands_OutsideOrders4(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.BuyBands = []BuyBand{BuyBand{Band{0.002344, 0.004689, 0.009378, 10.0, 40.0, 80.0, 0.0}}}
	buyOrders := []*Order{&Order{"DAIUSD", "BK01", 0, 0.990621, 50.0, 20.0, 1, "New", 0, 0, "1515755942"}}	//create outside-band bid order
	sellOrders := []*Order{}
	refPrice := 1.00					//set ref price of asset to 1
	assert.NotEmpty(t, bands.OutsideOrders(buyOrders, sellOrders, refPrice))				//assert order is out of boundary
	assert.Contains(t, bands.OutsideOrders(buyOrders, sellOrders, refPrice), buyOrders[0])	//assert order did not get corrupted
}


//Test if ask on boundary of minMargin is in-band
func Test_Bands_OutsideOrders5(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.SellBands = []SellBand{SellBand{Band{0.000428, 0.000856, 0.001711, 0.01, 0.1, 0.15, 0.0}}}
	sellOrders := []*Order{&Order{"DAIUSD", "BK01", 1, 1.000428, 50.0, 20.0, 1, "New", 0, 0, "1515755942"}}	//create ask order
	buyOrders := []*Order{}
	refPrice := 1.00					//set ref price of asset to 1
	assert.Empty(t, bands.OutsideOrders(buyOrders, sellOrders, refPrice))		//assert order is within boundary
}

//Test if ask on boundary of maxMargin is in-band
func Test_Bands_OutsideOrders6(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.SellBands = []SellBand{SellBand{Band{0.000428, 0.000856, 0.001711, 0.01, 0.1, 0.15, 0.0}}}
	sellOrders := []*Order{&Order{"DAIUSD", "BK02", 1, 1.001711, 10.0, 10.0, 1, "New", 0, 0, "1515755945"}}	//create ask order
	buyOrders := []*Order{}
	refPrice := 1.00					//set ref price of asset to 1
	assert.Empty(t, bands.OutsideOrders(buyOrders, sellOrders, refPrice))		//assert order is within boundary
}

//Test if ask on minMargin-- is in-band
func Test_Bands_OutsideOrders7(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.SellBands = []SellBand{SellBand{Band{0.000428, 0.000856, 0.001711, 0.01, 0.1, 0.15, 0.0}}}
	sellOrders := []*Order{&Order{"DAIUSD", "BK02", 1, 1.000427, 10.0, 10.0, 1, "New", 0, 0, "1515755945"}}	//create ask order
	buyOrders := []*Order{}
	refPrice := 1.00					//set ref price of asset to 1
	assert.NotEmpty(t, bands.OutsideOrders(buyOrders, sellOrders, refPrice))				//assert order is out of boundary
	assert.Contains(t, bands.OutsideOrders(buyOrders, sellOrders, refPrice), sellOrders[0])	//assert order did not get corrupted
}

//Test if ask on maxMargin++ is in-band
func Test_Bands_OutsideOrders8(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.SellBands = []SellBand{SellBand{Band{0.000428, 0.000856, 0.001711, 0.01, 0.1, 0.15, 0.0}}}
	sellOrders := []*Order{&Order{"DAIUSD", "BK02", 1, 1.001712, 10.0, 10.0, 1, "New", 0, 0, "1515755945"}}	//create ask order
	buyOrders := []*Order{}
	refPrice := 1.00					//set ref price of asset to 1
	assert.NotEmpty(t, bands.OutsideOrders(buyOrders, sellOrders, refPrice))				//assert order is out of boundary
	assert.Contains(t, bands.OutsideOrders(buyOrders, sellOrders, refPrice), sellOrders[0])	//assert order did not get corrupted
}

//BAND
//Test if band has improper parameters
func Test_Band_VerifyBand1(t *testing.T) {
	bands := new(Bands)								//create bands instance
	bands.LoadBands()								//load bands from bands.json
	assert.Nil(t, bands.BuyBands[0].VerifyBand())	//assert band is valid
}

//Test if band has improper parameters
func Test_Band_VerifyBand2(t *testing.T) {
	bands := new(Bands)								//create bands instance
	bands.LoadBands()								//load bands from bands.json
	band := &bands.BuyBands[0]						//get buy band
	band.MinMargin = band.AvgMargin + 0.000001		//set MinMargin to be AvgMargin++
	assert.Error(t, band.VerifyBand())				//assert band is invalid
}

//Test if band has improper parameters
func Test_Band_VerifyBand3(t *testing.T) {
	bands := new(Bands)								//create bands instance
	bands.LoadBands()								//load bands from bands.json
	band := &bands.BuyBands[0]						//get buy band
	band.AvgMargin = band.MaxMargin + 0.000001		//set AvgMargin to be MaxMargin++
	assert.Error(t, band.VerifyBand())				//assert band is invalid
}

//Test if band has improper parameters
func Test_Band_VerifyBand4(t *testing.T) {
	bands := new(Bands)								//create bands instance
	bands.LoadBands()								//load bands from bands.json
	band := &bands.BuyBands[0]						//get buy band
	band.MinMargin = band.MaxMargin					//set MinMargin to be MaxMargin
	assert.Error(t, band.VerifyBand())				//assert band is invalid
}

//Test if band has improper parameters
func Test_Band_VerifyBand5(t *testing.T) {
	bands := new(Bands)								//create bands instance
	bands.LoadBands()								//load bands from bands.json
	band := &bands.BuyBands[0]						//get buy band
	band.MinAmount = band.AvgAmount + 0.00001		//set MinAmount to be AvgAmount++
	assert.Error(t, band.VerifyBand())				//assert band is invalid
}

//Test if band has improper parameters
func Test_Band_VerifyBand6(t *testing.T) {
	bands := new(Bands)								//create bands instance
	bands.LoadBands()								//load bands from bands.json
	band := &bands.BuyBands[0]						//get buy band
	band.AvgAmount = band.MaxAmount + 0.00001		//set AvgAmount to be MaxAmount++
	assert.Error(t, band.VerifyBand())				//assert band is invalid
}

//Test if band has improper parameters
func Test_Band_VerifyBand7(t *testing.T) {
	bands := new(Bands)								//create bands instance
	bands.LoadBands()								//load bands from bands.json
	band := &bands.BuyBands[0]						//get buy band
	band.MinAmount = band.MaxAmount + 0.00001		//set MinAmount to be MaxAmount++
	assert.Error(t, band.VerifyBand())				//assert band is invalid
}

func Test_Band_ExecessiveOrders1(t *testing.T) {
	sBand := SellBand{Band{0.1, 0.15, 0.2, 4.0, 6.0, 8.0, 0.01}}	//create buy band
	//MinMargin - 0.1, MaxMargin = 0.2, MinAmount = 4, MaxAmount < 8
	targetPrice := 1.0												//set ref price to 8.5
	//With RefPrice of 1.0 -> MinPrice = 1.1 & MaxPrice = 1.2
	askOrders := []*Order{					 						//create orders
		&Order{"DAIUSD", "BK01", 1, 1.1, 14.13, 1, 1, "New", 0, 0, "1515755945"},	//Order in-band
		&Order{"DAIUSD", "BK02", 1, 1.12, 10.17, 2, 1, "New", 0, 0, "1515755945"},	//Order in-band
		&Order{"DAIUSD", "BK03", 1, 1.16, 11.84, 3, 1, "New", 0, 0, "1515755945"},	//Order in-band
		&Order{"DAIUSD", "BK04", 1, 1.20, 12.96, 4, 1, "New", 0, 0, "1515755945"},	//Order in-band
	}
	ordersToKill := sBand.ExcessiveOrders(askOrders, targetPrice)	//find which orders need to be cancelled to stay under band.MaxAmount
	assert.Contains(t, ordersToKill, askOrders[2])					//check that order BK03 was selected to be cancelled
	assert.Equal(t, len(ordersToKill), 1)							//check that no other orders were specified to be cancelled
}

func Test_Band_ExecessiveOrders2(t *testing.T) {
	bBand := BuyBand{Band{0.1, 0.11, 0.2, 4.0, 6.0, 7.0, 0.01}}	//create buy band
	//MinMargin - 0.1, MaxMargin = 0.2, MinAmount = 4, MaxAmount < 7
	targetPrice := 1.0													//set ref price to 1.0
	//With RefPrice of 1.0 -> MinPrice = 0.9 & MaxPrice = 0.8
	bidOrders := []*Order{												//create orders
		&Order{"DAIUSD", "BK01", 0, 0.8907, 14.13, 1, 1, "New", 0, 0, "1515755945"},	//Order in-band
		&Order{"DAIUSD", "BK02", 0, 0.8714, 10.17, 2, 1, "New", 0, 0, "1515755945"},	//Order in-band
		&Order{"DAIUSD", "BK03", 0, 0.8465, 11.84, 3, 1, "New", 0, 0, "1515755945"},	//Order in-band
		&Order{"DAIUSD", "BK04", 0, 0.8277, 12.96, 4, 1, "New", 0, 0, "1515755945"},	//Order in-band
	}
	ordersToKill := bBand.ExcessiveOrders(bidOrders, targetPrice)		//find which orders need to be cancelled to stay under band.MaxAmount
	assert.Contains(t, ordersToKill, bidOrders[3])						//check that order BK03 was selected to be cancelled
	assert.Equal(t, len(ordersToKill), 1)								//check that no other orders were specified to be cancelled
}

func Test_Band_ExecessiveOrders3(t *testing.T) {
	bBand := BuyBand{Band{0.01, 0.013, 0.02, 4.0, 6.0, 8.0, 0.01}}	//create buy band
	//MinMargin - 0.01, MaxMargin = 0.02, MinAmount = 4, MaxAmount < 8
	targetPrice := 1.0													//set ref price to 1.0
	//With RefPrice of 1.0 -> MinPrice = 0.9 & MaxPrice = 0.8
	bidOrders := []*Order{												//create orders
		&Order{"DAIUSD", "BK00", 0, 0.9952, 7.21, 1, 1, "New", 0, 0, "1515755945"},		//Order out-of-band
		&Order{"DAIUSD", "BK01", 0, 0.9899, 14.13, 1, 1, "New", 0, 0, "1515755945"},	//Order in-band
		&Order{"DAIUSD", "BK02", 0, 0.9877, 10.17, 2, 1, "New", 0, 0, "1515755945"},	//Order in-band
		&Order{"DAIUSD", "BK03", 0, 0.9832, 11.84, 3, 1, "New", 0, 0, "1515755945"},	//Order in-band
		&Order{"DAIUSD", "BK04", 0, 0.9801, 12.96, 4, 1, "New", 0, 0, "1515755945"},	//Order in-band
		&Order{"DAIUSD", "BK05", 0, 0.9762, 8.32, 1, 1, "New", 0, 0, "1515755945"},		//Order out-of-band
	}
	ordersToKill := bBand.ExcessiveOrders(bidOrders, targetPrice)		//find which orders need to be cancelled to stay under band.MaxAmount
	assert.Contains(t, ordersToKill, bidOrders[3])						//check that order BK03 was selected to be cancelled
	assert.Equal(t, len(ordersToKill), 1)								//check that no other orders were specified to be cancelled
}

//No return value, can't test individually
//This is tested as a dependancy in excessiveOrders
func Test_Band_GetAllCombinationsOfSizeN(t *testing.T) {
	bands := new(Bands)								//create bands instance
	bands.LoadBands() 								//load banks from bands.json
	band := &bands.BuyBands[0]						//get buy band
	askOrders := []*Order{							//create orders
		&Order{"DAIUSD", "BK01", 1, 1.000428, 14.13, 1, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK02", 1, 1.000502, 10.17, 2, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK03", 1, 1.000675, 11.84, 3, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK04", 1, 1.000788, 12.96, 4, 1, "New", 0, 0, "1515755945"},
	}
	for i := 1; i <= len(askOrders); i++ {
		band.GetAllCombinationsOfSizeN(askOrders, i)
	}
}

//No return value, can't test individually
//This is tested as a dependancy in excessiveOrders
func Test_Band_CombinationUtil(t *testing.T) {

}

//Test if band includes bid order with price at MinMargin
func Test_Band_Includes1(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.BuyBands[0]			//get buy band
	targetPrice := 1.00					//set ref price of asset to 1
	orderPrice := (1 - band.MinMargin) * targetPrice	//calculate bid price to be on exactly MinMargin boundary
	assert.True(t, band.Includes(orderPrice, 1.00))		//assert bid order is in boundary
}

//Test if band includes bid order with price at MinMargin++
func Test_Band_Includes2(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.BuyBands[0]			//get buy band
	targetPrice := 1.00					//set ref price of asset to 1
	orderPrice := (1 - band.MinMargin) * targetPrice + 0.000001	//calculate bid price to be just out of MinMargin boundary
	assert.False(t, band.Includes(orderPrice, targetPrice))		//assert bid order is out of boundary
}

//Test if band includes bid order with price at MaxMargin
func Test_Band_Includes3(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.BuyBands[0]			//get buy band
	targetPrice := 1.00					//set ref price of asset to 1
	orderPrice := (1 - band.MaxMargin) * targetPrice		//calculate bid price to be exactly MaxMargin boundary
	assert.True(t, band.Includes(orderPrice, targetPrice))	//assert bid order is in boundary
}

//Test if band includes bid order with price at MaxMargin--
func Test_Band_Includes4(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.BuyBands[0]			//get buy band
	targetPrice := 1.00					//set ref price of asset to 1
	orderPrice := (1 - band.MaxMargin) * targetPrice - 0.000001	//calculate bid price to be just out of MaxMargin boundary
	assert.False(t, band.Includes(orderPrice, targetPrice))		//assert bid order is out of boundary
}

//Test if band includes ask order with price at MinMargin
func Test_Band_Includes5(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.SellBands[0]			//get sell band
	targetPrice := 1.00					//set ref price of asset to 1
	orderPrice := (1 + band.MinMargin) * targetPrice	//calculate ask price to be exactly MinMargin boundary
	assert.True(t, band.Includes(orderPrice, 1.00))		//assert ask order is in boundary
}

//Test if band includes ask order with price at MinMargin--
func Test_Band_Includes6(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.SellBands[0]			//get sell band
	targetPrice := 1.00					//set ref price of asset to 1
	orderPrice := (1 + band.MinMargin) * targetPrice - 0.000001	//calculate ask price to be just out of MaxMargin boundary
	assert.False(t, band.Includes(orderPrice, targetPrice))		//assert ask order is out of boundary
}

//Test if band includes ask order with price at MaxMargin
func Test_Band_Includes7(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.SellBands[0]			//get sell band
	targetPrice := 1.00					//set ref price of asset to 1
	orderPrice := (1 + band.MaxMargin) * targetPrice		//calculate ask price to be exactly MaxMargin boundary
	assert.True(t, band.Includes(orderPrice, targetPrice))	//assert ask order is in boundary
}

//Test if band includes ask order with price at MaxMargin++
func Test_Band_Includes8(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.SellBands[0]			//get sell band
	targetPrice := 1.00					//set ref price of asset to 1
	orderPrice := (1 + band.MaxMargin) * targetPrice + 0.000001	//calculate ask price to be just out of MaxMargin boundary
	assert.False(t, band.Includes(orderPrice, targetPrice))		//assert ask order is out of boundary
}

func Test_Band_TotalAmount1(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	askOrders := []*Order{
		&Order{"DAIUSD", "BK01", 1, 1.000428, 14.13, 10.13, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK02", 1, 1.000502, 10.17, 10.17, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK03", 1, 1.000675, 11.84, 10.84, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK04", 1, 1.000788, 12.96, 10.96, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK05", 1, 1.000972, 19.92, 10.92, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK06", 1, 1.001002, 22.64, 10.64, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK07", 1, 1.001183, 11.81, 10.81, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK08", 1, 1.001599, 17.77, 10.77, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK09", 1, 1.001711, 15.08, 10.08, 1, "New", 0, 0, "1515755945"},
	}
	band := bands.BuyBands[0]			//get buy band
	assert.Equal(t, band.TotalAmount(askOrders), 95.32)	//assert total of all orders
}

func Test_Band_TotalAmount2(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	askOrders := []*Order{
		&Order{"DAIUSD", "BK01", 1, 1.000428, 14.13, 10.13, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK02", 1, 1.000502, 10.17, 10.17, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK03", 1, 1.000675, 11.84, 10.84, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK04", 1, 1.000788, 12.96, 10.96, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK05", 1, 1.000972, 19.92, 10.92, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK06", 1, 1.001002, 22.64, 10.64, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK07", 1, 1.001183, 11.81, 10.81, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK08", 1, 1.001599, 17.77, 10.77, 1, "New", 0, 0, "1515755945"},
		&Order{"DAIUSD", "BK09", 1, 1.001711, 15.08, 10.08, 1, "New", 0, 0, "1515755945"},
	}
	band := bands.SellBands[0]			//get sell band
	assert.Equal(t, band.TotalAmount(askOrders), 95.32)	//assert total of all orders
}

//BUY BAND
func Test_BuyBand_Includes1(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.BuyBands[0]			//get buy band
	targetPrice := 1.0 					//set ref price to 1.0
	orderPrice := (1 - band.MinMargin) * targetPrice			//calculate bid price to be on exactly MinMargin boundary
	assert.True(t, band.Includes(orderPrice, targetPrice))		//assert bid order is in boundary
}

func Test_BuyBand_Includes2(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.BuyBands[0]			//get buy band
	targetPrice := 1.0 					//set ref price to 1.0
	orderPrice := (1 - band.MinMargin) * targetPrice + 0.00001	//calculate bid price to be just out of  MinMargin++ boundary
	assert.False(t, band.Includes(orderPrice, targetPrice))		//assert bid order is in boundary
}

func Test_BuyBand_Includes3(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.BuyBands[0]			//get buy band
	targetPrice := 1.0 					//set ref price to 1.0
	orderPrice := (1 - band.MaxMargin) * targetPrice			//calculate bid price to be on exactly MaxMargin boundary
	assert.True(t, band.Includes(orderPrice, targetPrice))		//assert bid order is in boundary
}

func Test_BuyBand_Includes4(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.BuyBands[0]			//get buy band
	targetPrice := 1.0 					//set ref price to 1.0
	orderPrice := (1 - band.MaxMargin) * targetPrice - 0.00001	//calculate bid price to be just out of  MaxMargin-- boundary
	assert.False(t, band.Includes(orderPrice, targetPrice))		//assert bid order is in boundary
}


func Test_BuyBand_ApplyMargin1(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.BuyBands[0]			//get buy band
	targetPrice := 1.0					//set ref price to 1.0
	margin := band.MinMargin 			//set margin to MinMargin
	assert.Equal(t, band.ApplyMargin(targetPrice, margin), (1 - margin) * targetPrice)	//assert ApplyMargin equals adjusted margin
}

func Test_BuyBand_ApplyMargin2(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.BuyBands[0]			//get buy band
	targetPrice := 1.0					//set ref price to 1.0
	margin := band.MaxMargin 			//set margin to MaxMargin
	assert.Equal(t, band.ApplyMargin(targetPrice, margin), (1 - margin) * targetPrice)	//assert ApplyMargin equals adjusted margin
}

func Test_BuyBand_AvgPrice(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.BuyBands[0]			//get buy band
	targetPrice := 1.0					//set ref price to 1.0
	margin := band.AvgMargin 			//set margin to AvgMargin
	assert.Equal(t, band.ApplyMargin(targetPrice, margin), (1 - margin) * targetPrice)	//assert ApplyMargin equals adjusted margin
}

//SELL BAND
func Test_SellBand_Includes1(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.SellBands[0]			//get sell band
	targetPrice := 1.0 					//set ref price to 1.0
	orderPrice := (1 + band.MinMargin) * targetPrice			//calculate bid price to be on exactly MinMargin boundary
	assert.True(t, band.Includes(orderPrice, targetPrice))		//assert ask order is in boundary
}

func Test_SellBand_Includes2(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.SellBands[0]			//get sell band
	targetPrice := 1.0 					//set ref price to 1.0
	orderPrice := (1 + band.MinMargin) * targetPrice - 0.00001	//calculate bid price to be just out of  MinMargin-- boundary
	assert.False(t, band.Includes(orderPrice, targetPrice))		//assert ask order is in boundary
}

func Test_SellBand_Includes3(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.SellBands[0]			//get sell band
	targetPrice := 1.0 					//set ref price to 1.0
	orderPrice := (1 + band.MaxMargin) * targetPrice			//calculate bid price to be on exactly MaxMargin boundary
	assert.True(t, band.Includes(orderPrice, targetPrice))		//assert ask order is in boundary
}

func Test_SellBand_Includes4(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.SellBands[0]			//get sell band
	targetPrice := 1.0 					//set ref price to 1.0
	orderPrice := (1 + band.MaxMargin) * targetPrice + 0.00001	//calculate bid price to be just out of MinMargin boundary
	assert.False(t, band.Includes(orderPrice, targetPrice))		//assert ask order is in boundary
}

func Test_SellBand_ApplyMargin1(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.SellBands[0]			//get sell band
	targetPrice := 1.0					//set ref price to 1.0
	margin := band.MinMargin 			//set margin to MinMargin
	assert.Equal(t, band.ApplyMargin(targetPrice, margin), (1 + margin) * targetPrice)	//assert ApplyMargin equals adjusted margin
}

func Test_SellBand_ApplyMargin2(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.SellBands[0]			//get sell band
	targetPrice := 1.0					//set ref prie to 1.0
	margin := band.MaxMargin 			//set margin to MaxMargin
	assert.Equal(t, band.ApplyMargin(targetPrice, margin), (1 + margin) * targetPrice)	//assert ApplyMargin equals adjusted margin
}

func Test_SellBand_AvgPrice(t *testing.T) {
	bands := new(Bands)					//create bands instance
	bands.LoadBands()					//load bands from bands.json
	band := bands.SellBands[0]			//get sell band
	targetPrice := 1.0					//set ref price to 1.0
	margin := band.AvgMargin 			//set margin to AvgMargin
	assert.Equal(t, band.ApplyMargin(targetPrice, margin), (1 + margin) * targetPrice)	//assert ApplyMargin equals adjusted margin
}