package lib

import (
	"fmt"
)

// CurrentPrice is the return object containing Turtle's value in BTC and USD
type CurrentPrice struct {
	CurrentUsdPrice string `json:"usdPrice"`
	CurrentBtcPrice string `json:"btcPrice"`
	usdPrice        float64
	btcPrice        float64
}

// SetCurrentPrices is kind of a hacky way to set the strings in the struct so I don't have to mess with a custom map right now
func (price CurrentPrice) SetCurrentPrices() CurrentPrice {
	price.CurrentUsdPrice = fmt.Sprintf("$%.8f", price.usdPrice)
	price.CurrentBtcPrice = fmt.Sprintf("Ƀ%.8f", price.btcPrice)
	return price
}
