package lib

import (
	"github.com/pkg/errors"
)

// GetPriceHash is the main driver that will get both the BTC and USD prices
func GetPriceHash(forceCheck bool) (price CurrentPrice, err error) {
	currentTrtlBtcPrice, getBtcPriceErr := GetTrtlToBtcPrice(forceCheck)

	if getBtcPriceErr != nil {
		return price, errors.Wrap(getBtcPriceErr, "Problem getting BTC Price")
	}
	price.btcPrice = currentTrtlBtcPrice

	currentUsdBtcPrice, getUsdBtcErr := GetBtcToUsdPrice(forceCheck)

	if getUsdBtcErr != nil {
		return price, getUsdBtcErr
	}

	trtlToUsd := currentTrtlBtcPrice * currentUsdBtcPrice

	price.usdPrice = trtlToUsd
	price = price.SetCurrentPrices()

	return price, err
}

// GetTrtlToBtcPrice is the main turtle to bitcoin price check
// TODO: add redis cache check here in case we got the price recently
func GetTrtlToBtcPrice(forceCheck bool) (btcPrice float64, err error) {
	// TODO: add redis cache check here in case we got the price recently
	var cache float64

	if cache != 0 && !forceCheck {
		return cache, nil
	}

	return PullTrtlToBtcPrice()
}

// ConvertTurtle does the math for converting turtle coins into BTC and USD
func ConvertTurtle(trtl int64, forceCheck bool) (priceHash CurrentPrice, err error) {
	currentPrice, getCurrentPriceError := GetPriceHash(forceCheck)
	if getCurrentPriceError != nil {
		return priceHash, errors.Wrap(getCurrentPriceError, "Problem getting the current price")
	}
	currentPrice.btcPrice = currentPrice.btcPrice * float64(trtl)
	currentPrice.usdPrice = currentPrice.usdPrice * float64(trtl)
	convertedPrice := currentPrice.SetCurrentPrices()
	return convertedPrice, err
}
