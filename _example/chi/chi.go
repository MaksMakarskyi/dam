package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MaksMakarskyi/dam"
	"github.com/go-chi/chi/v5"
)

func SayHi(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hi")
}

func SayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello")
}

func main() {
	r := chi.NewRouter()

	r.Route("/greetings", func(r chi.Router) {
		// dam.Limit returns chi's middleware type, so no adapter is needed. The
		// routes below share this one limiter: 50 requests per second per IP
		// between them, not each. chi panics if Use comes after the routes.
		r.Use(dam.Limit(dam.NewFixedWindow(50, time.Second), dam.KeyByIP))

		r.Get("/hello", SayHello)

		// A route may stack a limiter of its own on top of the group's. nil
		// takes the package defaults, keyed by the token's "sub" claim.
		r.Get("/hi", dam.LimitFunc(nil, dam.KeyByJWTClaim("sub"), SayHi))
	})

	// A third limiter caps the service as a whole. Each layer needs its own
	// instance; reusing one would charge a single request against it twice.
	globalLimiter := dam.NewFixedWindow(100, time.Second)

	server := http.Server{
		Addr: ":8080",
		// A chi.Router is an http.Handler, and dam.KeyGlobal keys every request
		// alike, so this is one service-wide budget over the whole router.
		Handler: dam.LimitHandler(globalLimiter, dam.KeyGlobal, r),
	}

	log.Fatal(server.ListenAndServe())
}
