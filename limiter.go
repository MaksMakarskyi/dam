package dam

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type KeyFunc func(r *http.Request) (string, error)
type Middleware func(next http.Handler) http.Handler

type Limiter interface {
	ShouldThrottle(r *http.Request) (bool, error)
}

func LimitFixedWindow(reqLimit uint64, windowLen time.Duration, keyFn KeyFunc) Middleware {
	return func(next http.Handler) http.Handler {
		fwl := &FixedWindowLimiter{
			mu:        sync.Mutex{},
			reqLimit:  reqLimit,
			windowLen: windowLen,
			keyFn:     keyFn,
			counter:   make(map[string]*Period),
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			throttle, err := fwl.ShouldThrottle(r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if throttle {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type FixedWindowLimiter struct {
	mu        sync.Mutex
	reqLimit  uint64
	windowLen time.Duration
	keyFn     KeyFunc
	counter   map[string]*Period
}

func (fwl *FixedWindowLimiter) ShouldThrottle(r *http.Request) (bool, error) {
	key, err := fwl.keyFn(r)
	if err != nil {
		return false, fmt.Errorf("failed to get the key from request: %w", err)
	}

	fwl.mu.Lock()
	defer fwl.mu.Unlock()

	now := time.Now()
	period, ok := fwl.counter[key]
	if !ok || period.end.Before(now) {
		period = &Period{
			start: now,
			end:   now.Add(fwl.windowLen),
			count: 0,
		}
		fwl.counter[key] = period
	}

	period.count += 1
	if period.count > fwl.reqLimit {
		return true, nil
	}

	return false, nil
}

type Period struct {
	start time.Time
	end   time.Time
	count uint64
}
