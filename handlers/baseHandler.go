package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	lib "github.com/y4htse/turtle-utils/lib"
)

// BaseHandler is the health check function
func BaseHandler(c *gin.Context) {
	price, err := lib.GetPriceHash(true)
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
