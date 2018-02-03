package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	lib "github.com/y4htse/turtle-utils/lib"
)

// BaseHandler is the health check function
func BaseHandler(c *gin.Context) {
	trtl := c.DefaultQuery("trtl", "1")

	trtlInt, intConvErr := strconv.ParseInt(trtl, 10, 64)

	if intConvErr != nil {
		errHandler(intConvErr, c)
		return
	}
	price, err := lib.ConvertTurtle(trtlInt, false)
	if err != nil {
		errHandler(err, c)
		return
	}
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"trtl":     trtlInt,
		"usdPrice": price.CurrentUsdPrice,
		"btcPrice": price.CurrentBtcPrice,
	})

}

func errHandler(err error, c *gin.Context) {
	c.HTML(http.StatusInternalServerError, "500.tmpl", gin.H{
		"error": err,
	})
}
