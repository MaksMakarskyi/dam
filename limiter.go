package dam

import (
	"context"
	"time"
)

// Limiter decides whether a request may proceed. Each distinct key carries its
// own budget, so what a key stands for — an IP address, an API key, a user —
// is what the limit applies to.
//
// Limit both reports the decision and records the request, so calling it twice
// spends two units of budget. Implementations must be safe for concurrent use.
type Limiter interface {
	Limit(ctx context.Context, key string) (LimitResult, error)
}

// LimitResult is the outcome of one [Limiter.Limit] call. It carries enough to
// answer the request and to fill in the RateLimit response headers.
type LimitResult struct {
	// Allowed reports whether the request may proceed.
	Allowed bool

	// Limit is the ceiling in force for this key.
	Limit int

	// Remaining is how much budget is left after this request, never negative.
	Remaining int

	// ResetAt is when the budget returns to Limit.
	ResetAt time.Time

	// RetryAfter is how long until the budget returns, meaningful when Allowed
	// is false.
	RetryAfter time.Duration
}
