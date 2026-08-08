package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MaksMakarskyi/dam"
	"github.com/MaksMakarskyi/dam/keyfunc"
	"github.com/MaksMakarskyi/dam/limiter"
)

func SayHi(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hi")
}

func SayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello")
}

func main() {
	mux := http.NewServeMux()

	// A limiter of its own gives this route a budget nothing else draws from.
	localLimiter := limiter.NewFixedWindowLimiter(10, time.Second)

	// Wrap a func(http.ResponseWriter, *http.Request) with dam.LimitFunc.
	// keyfunc.JWTClaim reads the "sub" claim of the bearer token, so each
	// authenticated user gets their own 10 requests per second.
	mux.HandleFunc("/hi", dam.LimitFunc(localLimiter, keyfunc.JWTClaim("sub"), SayHi))

	// Passing nil for either parameter takes the package defaults: dam.DefaultLimit
	// requests per dam.DefaultWindow, counted into one bucket shared by every
	// client rather than per client.
	mux.HandleFunc("/hello", dam.LimitFunc(nil, nil, SayHello))

	// A second limiter caps the service as a whole. It must be a separate
	// instance: sharing one with the routes above would charge each request
	// twice and halve both limits.
	globalLimiter := limiter.NewFixedWindowLimiter(100, time.Second)

	server := http.Server{
		Addr: ":8080",
		// dam.LimitHandler wraps an http.Handler, here the whole mux.
		// keyfunc.Common keys every request identically, so the 100 per second
		// is a single service-wide budget. Requests to /hi and /hello need room
		// in this limiter and in their own.
		Handler: dam.LimitHandler(globalLimiter, keyfunc.Common, mux),
	}

	log.Fatal(server.ListenAndServe())
}
