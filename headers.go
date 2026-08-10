package dam

import (
	"math"
	"net/http"
	"strconv"
	"time"
)

// SetRateLimitHeaders writes the RateLimit-Limit, RateLimit-Remaining and
// RateLimit-Reset fields describing result. RateLimit-Reset is a whole number
// of seconds from now, rounded up, as the IETF RateLimit header fields draft
// specifies — not a date.
//
// Every response carries these, allowed or not, so a client can pace itself
// before it is ever rejected. Call it before writing the status or body:
// headers set afterwards are discarded.
//
// It is exported for adapters targeting frameworks this module does not ship,
// which need the same fields on their own response type. Handlers wrapped by
// [Limit], [LimitHandler] or [LimitFunc] already get them.
func SetRateLimitHeaders(w http.ResponseWriter, result LimitResult) {
	w.Header().Set("RateLimit-Limit", strconv.Itoa(result.Limit))
	w.Header().Set("RateLimit-Remaining", strconv.Itoa(result.Remaining))
	w.Header().Set("RateLimit-Reset", strconv.Itoa(resetSeconds(result.ResetAt)))
}

// SetRetryAfter writes the Retry-After field telling the client how long to
// wait, in seconds, as RFC 9110 permits. It belongs only on a rejection:
// pair it with [SetRateLimitHeaders] when result.Allowed is false, and leave it
// off otherwise.
//
// The value is derived from result.ResetAt, which for a fixed window is when
// the whole budget returns. A limiter that refills gradually will want to
// advertise its next single unit instead.
func SetRetryAfter(w http.ResponseWriter, result LimitResult) {
	w.Header().Set("Retry-After", strconv.Itoa(resetSeconds(result.ResetAt)))
}

// resetSeconds is the whole number of seconds until resetAt, rounded up so a
// client that waits that long is never early. It never goes below zero: a reset
// already in the past means the budget is available now.
func resetSeconds(resetAt time.Time) int {
	return max(0, int(math.Ceil(time.Until(resetAt).Seconds())))
}
