package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/thinkerou/favicon"
	handlers "github.com/y4htse/turtle-utils/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalln("Must set $PORT")
	}
	r := gin.Default()
	r.Use(favicon.New("favicon.ico"))
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		handlers.BaseHandler(c)
	})
	r.GET("/price", func(c *gin.Context) {
		handlers.PriceHandler(c)
	})
	r.GET("/convert", func(c *gin.Context) {
		handlers.ConvertHandler(c)
	})
	r.Run(fmt.Sprintf("0.0.0.0:%s", port))
}
