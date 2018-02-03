package lib

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// PullBtcToUsdPrice hits the Coinbase API to get the current Bitcoin to USD price
func PullBtcToUsdPrice() (usdPrice float64, err error) {
	btcToUsdURL := "https://api.coinbase.com/v2/prices/BTC-USD/spot"
	type data struct {
		Amount string `json:"amount"`
	}
	type coinbaseResult struct {
		Data data `json:"data"`
	}
	// {"data":{"base":"BTC","currency":"USD","amount":"11110.66"}}
	type tradeOgreResult struct {
		// Number int `json:"number"`
		Price string `json:"price"`
	}

	spaceClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	trtlToBtcReq, reqErr := http.NewRequest(http.MethodGet, btcToUsdURL, nil)

	if reqErr != nil {
		return usdPrice, reqErr
	}

	res, getErr := spaceClient.Do(trtlToBtcReq)
	if getErr != nil {
		return usdPrice, getErr
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return usdPrice, readErr
	}

	tradeOgreRes := coinbaseResult{}
	jsonErr := json.Unmarshal(body, &tradeOgreRes)
	if jsonErr != nil {
		return usdPrice, jsonErr
	}

	usdPrice, err = strconv.ParseFloat(tradeOgreRes.Data.Amount, 64)
	return usdPrice, err
}

// GetBtcToUsdPrice is the main bitcoin price check
// TODO: add redis cache check here in case we got the price recently
func GetBtcToUsdPrice(forceCheck bool) (usdPrice float64, err error) {
	var cache float64

	if cache != 0 && !forceCheck {
		return cache, nil
	}

	return PullBtcToUsdPrice()
}
