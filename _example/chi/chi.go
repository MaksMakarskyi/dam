package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MaksMakarskyi/dam"
	"github.com/MaksMakarskyi/dam/keyfunc"
	"github.com/MaksMakarskyi/dam/limiter"
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
		// dam.Limit returns func(http.Handler) http.Handler, which is exactly
		// chi's middleware type, so no adapter is needed. One limiter is built
		// here and shared by every route in the group: together they get 50
		// requests per second per IP address, not 50 each.
		//
		// chi panics if middleware is registered after routes, so Use comes first.
		r.Use(dam.Limit(limiter.NewFixedWindowLimiter(50, time.Second), keyfunc.IP))

		r.Get("/hello", SayHello)

		// A route may add a limiter of its own on top of the group's. Passing
		// nil takes the package default of dam.DefaultLimit requests per
		// dam.DefaultWindow, here keyed by the "sub" claim of the bearer token
		// so the budget is per authenticated user.
		r.Get("/hi", dam.LimitFunc(nil, keyfunc.JWTClaim("sub"), SayHi))
	})

	// A third limiter caps the service as a whole. Each layer needs its own
	// instance: reusing one would charge a single request against it twice.
	globalLimiter := limiter.NewFixedWindowLimiter(100, time.Second)

	server := http.Server{
		Addr: ":8080",
		// dam.LimitHandler wraps an http.Handler, and a chi.Router is one.
		// keyfunc.Common keys every request identically, making this a single
		// service-wide budget. A request to /greetings/hi passes all three
		// limiters and needs room in each.
		Handler: dam.LimitHandler(globalLimiter, keyfunc.Common, r),
	}

	log.Fatal(server.ListenAndServe())
}
