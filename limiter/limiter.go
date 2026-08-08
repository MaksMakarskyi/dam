package limiter

import (
	"context"
	"net/http"
	"time"
)

type KeyFunc func(r *http.Request) (string, error)

type Limiter interface {
	Limit(ctx context.Context, key string) (Result, error)
}

type Result struct {
	Allowed    bool
	Limit      int
	Remaining  int
	ResetAt    time.Time
	RetryAfter time.Duration
}
