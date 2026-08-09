package dam

import (
	"math"
	"net/http"
	"strconv"
	"time"
)

func setHeaders(w http.ResponseWriter, result LimitResult) {
	w.Header().Set("RateLimit-Limit", strconv.Itoa(result.Limit))
	w.Header().Set("RateLimit-Remaining", strconv.Itoa(result.Remaining))
	w.Header().Set("RateLimit-Reset", strconv.Itoa(resetSeconds(result.ResetAt)))
}

func setLimitedHeaders(w http.ResponseWriter, result LimitResult) {
	w.Header().Set("Retry-After", strconv.Itoa(resetSeconds(result.ResetAt)))
}

func resetSeconds(resetAt time.Time) int {
	return max(0, int(math.Ceil(time.Until(resetAt).Seconds())))
}
