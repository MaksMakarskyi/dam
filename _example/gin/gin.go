package main

import (
	"log"
	"net/http"
	"time"

	"github.com/MaksMakarskyi/dam"
	"github.com/MaksMakarskyi/dam/damgin"
	"github.com/gin-gonic/gin"
)

func SayHi(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hi!"})
}

func SayHello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello!"})
}

func main() {
	router := gin.Default()

	// damgin.Limit returns a gin.HandlerFunc, so it registers like any other
	// middleware. Here it caps the whole router at 100 requests per second per IP.
	globalLimiter := dam.NewFixedWindow(100, time.Second)
	router.Use(damgin.Limit(globalLimiter, dam.KeyByIP))

	// A route may stack a limiter of its own on top. Each layer needs its own
	// instance; reusing one would charge a single request against it twice.
	localLimiter := dam.NewFixedWindow(10, time.Second)
	router.GET("/hi", damgin.LimitFunc(localLimiter, dam.KeyByJWTClaim("sub"), SayHi))

	// nil, nil takes the package defaults keyed globally: one shared budget for
	// every caller, not a budget each.
	router.GET("/hello", damgin.LimitFunc(nil, nil, SayHello))

	log.Fatal(router.Run(":8080"))
}
