package dam

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/MaksMakarskyi/dam/limiter"
)

func setHeaders(w http.ResponseWriter, result limiter.Result) {
	w.Header().Set("RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
	w.Header().Set("RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
	w.Header().Set("RateLimit-Reset", result.ResetAt.UTC().Format(time.RFC1123))
}

func setLimitedHeaders(w http.ResponseWriter, result limiter.Result) {
	w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(result.RetryAfter.Seconds()))))
}
