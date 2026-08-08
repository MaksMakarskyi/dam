package middleware

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/MaksMakarskyi/dam/limiter"
)

type Middleware func(next http.Handler) http.Handler

type LimitHeader string

const (
	RetryAfter         LimitHeader = "Retry-After"
	ContentType        LimitHeader = "Content-Type"
	RateLimitLimit     LimitHeader = "RateLimit-Limit"
	RateLimitRemaining LimitHeader = "RateLimit-Remaining"
	RateLimitReset     LimitHeader = "RateLimit-Reset"
)

func LimitFixedWindow(reqLimit int, windowLen time.Duration, keyFn limiter.KeyFunc) Middleware {
	return func(next http.Handler) http.Handler {
		fwl := limiter.NewFixedWindowLimiter(reqLimit, windowLen)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, err := keyFn(r)
			if err != nil {
				http.Error(w, "failed to get a rate limiting key for request", http.StatusInternalServerError)
				return
			}

			result, err := fwl.Limit(r.Context(), key)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			fmt.Printf("%v\n", result.Allowed)

			SetHeaders(w, result)
			if !result.Allowed {
				SetLimitedHeaders(w, result)
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func SetHeaders(w http.ResponseWriter, result limiter.Result) {
	w.Header().Set(string(RateLimitLimit), fmt.Sprintf("%d", result.Limit))
	w.Header().Set(string(RateLimitRemaining), fmt.Sprintf("%d", result.Remaining))
	w.Header().Set(string(RateLimitReset), result.ResetAt.UTC().Format(time.RFC1123))
}

func SetLimitedHeaders(w http.ResponseWriter, result limiter.Result) {
	w.Header().Set(string(RetryAfter), strconv.Itoa(int(math.Ceil(result.RetryAfter.Seconds()))))
}
