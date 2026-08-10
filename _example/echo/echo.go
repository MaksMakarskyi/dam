package main

import (
	"net/http"
	"time"

	"github.com/MaksMakarskyi/dam"
	"github.com/MaksMakarskyi/dam/damecho"
	"github.com/labstack/echo/v5"
)

func SayHi(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Hi!"})
}

func SayHello(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Hello!"})
}

func main() {
	e := echo.New()

	// damecho.Limit returns echo.MiddlewareFunc, so it registers like any other
	// middleware. Here it caps the whole router at 100 requests per second per IP.
	globalLimiter := dam.NewFixedWindow(100, time.Second)
	e.Use(damecho.Limit(globalLimiter, dam.KeyByIP))

	// A route may stack a limiter of its own on top. Each layer needs its own
	// instance; reusing one would charge a single request against it twice.
	localLimiter := dam.NewFixedWindow(10, time.Second)
	e.GET("/hi", damecho.LimitFunc(localLimiter, dam.KeyByJWTClaim("sub"), SayHi))

	// nil, nil takes the package defaults keyed globally: one shared budget for
	// every caller, not a budget each.
	e.GET("/hello", damecho.LimitFunc(nil, nil, SayHello))

	if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
		e.Logger.Error("failed to start server", "error", err)
	}
}
