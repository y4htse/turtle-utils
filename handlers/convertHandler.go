package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	lib "github.com/y4htse/turtle-utils/lib"
)

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

	trtlValue, trtlConvertError := lib.ConvertTurtle(trtlInt, forcedBool)

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
