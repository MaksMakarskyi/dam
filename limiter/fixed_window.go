package limiter

import (
	"context"
	"sync"
	"time"
)

var _ Limiter = (*FixedWindowLimiter)(nil)

type FixedWindowLimiter struct {
	mu        sync.Mutex
	reqLimit  int
	windowLen time.Duration
	counter   map[string]*FixedWindow
}

func NewFixedWindowLimiter(
	reqLimit int,
	windowLen time.Duration,
) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		reqLimit:  reqLimit,
		windowLen: windowLen,
		counter:   make(map[string]*FixedWindow),
	}
}

func (fwl *FixedWindowLimiter) Limit(ctx context.Context, key string) (Result, error) {
	fwl.mu.Lock()
	defer fwl.mu.Unlock()

	now := time.Now()
	period, ok := fwl.counter[key]
	if !ok || !now.Before(period.end) {
		period = &FixedWindow{
			end:   now.Add(fwl.windowLen),
			count: 0,
		}
		fwl.counter[key] = period
	}

	period.count += 1

	res := Result{
		Allowed:    period.count <= fwl.reqLimit,
		Limit:      fwl.reqLimit,
		ResetAt:    period.end,
		RetryAfter: period.end.Sub(now),
	}

	if res.Allowed {
		res.Remaining = fwl.reqLimit - period.count
	} else {
		res.Remaining = 0
	}

	return res, nil
}

type FixedWindow struct {
	end   time.Time
	count int
}
