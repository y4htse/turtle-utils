package lib

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// PullTrtlToBtcPrice hits the tradeogre API to get the current Turtle to Bitcoin price
func PullTrtlToBtcPrice() (btcPrice float64, err error) {
	trtlToBtcURL := "https://tradeogre.com/api/v1/ticker/BTC-TRTL"
	// {"initialprice":"0.00000010","price":"0.00000016","high":"0.00000016","low":"0.00000006","volume":"17.18630467"}
	type tradeOgreResult struct {
		// Number int `json:"number"`
		Price string `json:"price"`
	}

	client := http.Client{
		Timeout: time.Second * 3, // Maximum of 3 secs
	}

	trtlToBtcReq, reqErr := http.NewRequest(http.MethodGet, trtlToBtcURL, nil)

	if reqErr != nil {
		return btcPrice, reqErr
	}

	res, getErr := client.Do(trtlToBtcReq)
	if getErr != nil {
		return btcPrice, getErr
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return btcPrice, readErr
	}

	tradeOgreRes := tradeOgreResult{}
	jsonErr := json.Unmarshal(body, &tradeOgreRes)
	if jsonErr != nil {
		return btcPrice, jsonErr
	}

	btcPrice, err = strconv.ParseFloat(tradeOgreRes.Price, 64)
	return btcPrice, err
}
