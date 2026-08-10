package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MaksMakarskyi/dam"
)

func SayHi(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hi")
}

func SayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello")
}

func main() {
	mux := http.NewServeMux()

	// A limiter of its own gives this route a budget nothing else draws from,
	// here 10 requests per second per authenticated user.
	localLimiter := dam.NewFixedWindow(10, time.Second)
	mux.HandleFunc("/hi", dam.LimitFunc(localLimiter, dam.KeyByJWTClaim("sub"), SayHi))

	// nil, nil takes the package defaults keyed globally: one shared budget for
	// every caller, not a budget each.
	mux.HandleFunc("/hello", dam.LimitFunc(nil, nil, SayHello))

	// A third limiter caps the server as a whole. Each layer needs its own
	// instance; reusing one would charge a single request against it twice.
	globalLimiter := dam.NewFixedWindow(100, time.Second)

	server := http.Server{
		Addr:    ":8080",
		Handler: dam.LimitHandler(globalLimiter, dam.KeyByIP, mux),
	}

	log.Fatal(server.ListenAndServe())
}
