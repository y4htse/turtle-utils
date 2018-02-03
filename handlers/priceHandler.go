package handlers

import (
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	lib "github.com/y4htse/turtle-utils/lib"
)

// PriceHandler is the function that will get the current trading prices for TurtleCoin
func PriceHandler(c *gin.Context) {
	forceCheck := c.DefaultQuery("force", "false")

	forcedBool, parseBoolErr := strconv.ParseBool(strings.ToUpper(forceCheck))
	if parseBoolErr != nil {
		forcedBool = false
	}

	log.Printf("Forced - %t\n", forcedBool)
	price, err := lib.GetPriceHash(forcedBool)
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
