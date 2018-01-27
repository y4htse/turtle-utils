package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/errors"
)

// BaseHandler is the health check function
func BaseHandler(c *gin.Context) {
	// TODO: add db and redis connection check
	c.JSON(200, gin.H{
		"status": "OK",
	})
}

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

// PriceHandler is the function that will get the current trading prices for TurtleCoin
func PriceHandler(c *gin.Context) {
	forceCheck := c.DefaultQuery("force", "false")

	forcedBool, parseBoolErr := strconv.ParseBool(strings.ToUpper(forceCheck))
	if parseBoolErr != nil {
		forcedBool = false
	}

	log.Printf("Forced - %t\n", forcedBool)
	price, err := GetPriceHash(forcedBool)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Problem getting the price of turtle in bitcoin",
		})
	} else {
		c.JSON(200, gin.H{
			"price": price,
		})
	}
}

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

// ConvertHandler will convert turtle coin (ints only) to BTC and USD
func ConvertHandler(c *gin.Context) {
	trtl := c.DefaultQuery("trtl", "1")

	trtlInt, intConvErr := strconv.ParseInt(trtl, 10, 64)

	if intConvErr != nil {
		c.JSON(404, gin.H{
			"error": fmt.Sprintf("Problem converting %s to an int", trtl),
		})
	}

	forceCheck := c.DefaultQuery("force", "false")

	forcedBool, parseBoolErr := strconv.ParseBool(strings.ToUpper(forceCheck))
	if parseBoolErr != nil {
		forcedBool = false
	}

	log.Printf("Forced - %t\n", forcedBool)

	trtlValue, trtlConvertError := ConvertTurtle(trtlInt, forcedBool)

	if trtlConvertError != nil {
		c.JSON(500, gin.H{
			"errors": errors.Wrap(trtlConvertError, "Could not convert at this time"),
		})
	} else {
		c.JSON(200, gin.H{
			"price": trtlValue,
		})
	}

}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalln("Must set $PORT")
	}
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		BaseHandler(c)
	})
	r.GET("/price", func(c *gin.Context) {
		PriceHandler(c)
	})
	r.GET("/convert", func(c *gin.Context) {
		ConvertHandler(c)
	})
	r.Run(fmt.Sprintf("0.0.0.0:%s", port))
}
