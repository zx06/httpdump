package main

import (
	"flag"
	"fmt"
	"log"
	"net/http/httputil"

	"github.com/gin-gonic/gin"
)

var port = "8080"

func init() {
	gin.SetMode(gin.ReleaseMode)
	flag.StringVar(&port, "p", "8080", "port to listen on")
}

func handler(c *gin.Context) {
	d, err := httputil.DumpRequest(c.Request, true)
	if err != nil {
		msg := fmt.Sprintf("couldn't dump request: %s", err)
		log.Printf(msg)
		c.String(500, "%s", err)
		return
	}
	b := string(d)
	log.Printf("\nrequest received:\n%s\n", b)
	c.String(200, b)
}

func main() {
	flag.Parse()
	router := gin.Default()
	router.Any("/", handler)
	router.Run(":" + port)
}
