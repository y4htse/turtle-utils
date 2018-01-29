package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	price, err := GetPriceHash(true)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.tmpl", gin.H{
			"error": err,
		})
	} else {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"usdPrice": price.CurrentUsdPrice,
			"btcPrice": price.CurrentBtcPrice,
		})
	}

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

// ThresholdRequestQueryObject represents a comparison that needs to be made
type ThresholdRequestQueryObject struct {
	Currency     string  `json:"currency"`
	Amount       float64 `json:"amount"`
	queryObjType QueryObjectType
}

// ThresholdRequest is what is POST'd to check to see if a threshold has been met
type ThresholdRequest struct {
	GreaterThan      ThresholdRequestQueryObject `json:"greaterThan,omitempty"`
	CallbackEndpoint string                      `json:"callbackEndpoint"`
	LessThan         ThresholdRequestQueryObject `json:"lessThan,omitempty"`
}

// QueryObjectType This is how we do the comparison
type QueryObjectType int

// This is how we compare
const (
	GreaterThan QueryObjectType = 1
	LessThan    QueryObjectType = 2
)

// SuccessCallbackEndpoint gets the URL for the success callback
func (req *ThresholdRequest) SuccessCallbackEndpoint() string {
	return fmt.Sprintf("%s/success", req.CallbackEndpoint)
}

// FailureCallbackEndpoint gets the URL for the failure callback
func (req *ThresholdRequest) FailureCallbackEndpoint() string {
	return fmt.Sprintf("%s/fail", req.CallbackEndpoint)
}

// GetCallbackUrls will conver the Urls
func (req *ThresholdRequest) GetCallbackUrls() (successURL *url.URL, failureURL *url.URL, err error) {
	successURL, successURLParseErr := url.ParseRequestURI(req.SuccessCallbackEndpoint())
	if successURLParseErr != nil {
		return successURL, failureURL, errors.Wrap(successURLParseErr, "Invalid successURL")
	}

	failureURL, failureURLParseErr := url.ParseRequestURI(req.FailureCallbackEndpoint())
	if failureURLParseErr != nil {
		return successURL, failureURL, errors.Wrap(successURLParseErr, "Invalid failureURL")
	}

	return successURL, failureURL, err
}

// Test is where the actual comparison happen
func (query *ThresholdRequestQueryObject) Test(currentAmount float64) (thresholdIsMet bool) {
	fmt.Printf("amount - %s\n", strconv.FormatFloat(query.Amount, 'f', -1, 64))
	fmt.Printf("currentAmount - %s\n", strconv.FormatFloat(currentAmount, 'f', -1, 64))
	q, _ := json.Marshal(query)
	fmt.Println(string(q))
	if query.Amount > currentAmount && query.queryObjType == GreaterThan {
		fmt.Println("Testing GreaterThan")
		return true
	}

	if query.Amount < currentAmount && query.queryObjType == LessThan {
		fmt.Println("Testing LessThan")
		return true
	}

	return thresholdIsMet
}

// PostFailureToCallback will attempt to call the user's failure callback service
func (req *ThresholdRequest) PostFailureToCallback(err error) {
	fmt.Println(err)
}

// PostSuccessToCallback is what will call the callback with the current price
func (req *ThresholdRequest) PostSuccessToCallback(price *CurrentPrice) {
	fmt.Println("Posting to success endpoint")
	url := req.SuccessCallbackEndpoint()

	binJSON, jsonMarshallErr := json.Marshal(gin.H{"currentPrice": price})
	if jsonMarshallErr != nil {
		req.PostFailureToCallback(jsonMarshallErr)
		return // Escape
	}
	fmt.Println(string(binJSON))
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(binJSON))
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		req.PostFailureToCallback(err)
		return
	}
	r, _ := json.Marshal(resp)
	fmt.Println(string(r))
	defer resp.Body.Close()
}

// TestThreshold This is what will call the apis if the threshold is met
func (req *ThresholdRequest) TestThreshold(successURL *url.URL, failureURL *url.URL) {
	fmt.Println("Testing the threshold")

	sucessURL, failureURL, callbackURLParseErr := req.GetCallbackUrls()
	if callbackURLParseErr != nil {
		fmt.Printf("Invalid callbackURL\n", callbackURLParseErr)
	}
	fmt.Printf("successUrl = %s\n", sucessURL)
	fmt.Printf("successUrl = %s\n", failureURL)

	if (req.GreaterThan.Amount <= 0 && req.LessThan.Amount <= 0) || (req.GreaterThan.Amount > 0 && req.LessThan.Amount > 0) {
		fmt.Println("SEND ERROR") // TODO
	}

	if (req.GreaterThan.Currency == "" && req.LessThan.Currency == "") || (req.GreaterThan.Currency != "" && req.LessThan.Currency != "") {
		fmt.Println("SEND ERROR") // TODO
	}

	var query *ThresholdRequestQueryObject
	if req.GreaterThan.Amount > 0 {
		query = &req.GreaterThan
		query.queryObjType = GreaterThan
		fmt.Println("It is GreaterThan!")
	} else if req.LessThan.Amount < 1 {
		query = &req.LessThan
		query.queryObjType = LessThan
		fmt.Println("It is LessThan!")
	} else {
		fmt.Println("WTF is happening")
	}

	q, _ := json.Marshal(query)
	fmt.Println(string(q))

	price, getPriceErr := GetPriceHash(false) // TODO: get the current price
	if getPriceErr != nil {
		req.PostFailureToCallback(getPriceErr)
		return
	}

	var currentPrice float64
	if query.Currency == "BTC" {
		currentPrice = price.btcPrice
	} else if query.Currency == "USD" {
		currentPrice = price.usdPrice
	} else {
		req.PostFailureToCallback(errors.New("Could not determine currency to compare"))
		return
	}
	fmt.Printf("currentPrice - %f\n", currentPrice)
	thresholdMet := query.Test(currentPrice)

	if thresholdMet {
		// TODO: call the API
		fmt.Println("THRESHOLD IS MET")
		req.PostSuccessToCallback(&price)
	} else {
		fmt.Println("THRESHOLD NOT MET")
	}

	fmt.Println("Done testing threshold")
}

// ThresholdHandler will check the price of TRTL against a given threshold and callback to an endpoint if TRTL is above that threshold
func ThresholdHandler(c *gin.Context) {
	var req *ThresholdRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.Wrap(bindErr, "Did not recognize request, please consult docs"),
		})
	}
	successCallbackEndoint, failureCallbackEndpoint, urlParseErr := req.GetCallbackUrls()
	if urlParseErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": urlParseErr,
		})
	}
	go req.TestThreshold(successCallbackEndoint, failureCallbackEndpoint)
	c.JSON(http.StatusOK, gin.H{
		"successCallbackEndpoint": successCallbackEndoint.String(),
		"failureCallbackEndpoint": failureCallbackEndpoint.String(),
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalln("Must set $PORT")
	}
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		BaseHandler(c)
	})
	r.GET("/price", func(c *gin.Context) {
		PriceHandler(c)
	})
	r.GET("/convert", func(c *gin.Context) {
		ConvertHandler(c)
	})
	r.POST("/threshold", func(c *gin.Context) {
		ThresholdHandler(c)
	})
	r.Run(fmt.Sprintf("0.0.0.0:%s", port))
}
