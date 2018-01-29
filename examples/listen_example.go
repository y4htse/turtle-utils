package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// ListenHandler is an example listener
func ListenHandler(c *gin.Context) {
	var req *gin.H
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		fmt.Printf("Error - %s", errors.Wrap(bindErr, "Did not recognize request, please consult docs"))
	}
	r, _ := json.Marshal(req)
	fmt.Println("Here it goes")
	fmt.Println(string(r))
	c.JSON(http.StatusOK, gin.H{
		"status": "Received",
	})
}

func main() {
	port := os.Getenv("LISTEN_PORT")
	if port == "" {
		log.Fatalln("Must set $LISTEN_PORT")
	}
	r := gin.Default()
	r.GET("/success", func(c *gin.Context) {
		fmt.Println("SUCCESS")
		ListenHandler(c)
	})
	r.GET("/failure", func(c *gin.Context) {
		fmt.Println("FAILURE")
		ListenHandler(c)
	})
	fmt.Printf("Listening on %s\n", port)
	r.Run(fmt.Sprintf("0.0.0.0:%s", port))
}
