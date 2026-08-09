package dam

import (
	"context"
	"sync"
	"time"
)

var _ Limiter = (*FixedWindow)(nil)

// FixedWindow allows reqLimit requests per key in each windowLen-long
// window. A key's first request opens its window; the count resets the moment
// that window expires.
//
// The reset is abrupt rather than gradual, so a client can spend its whole
// budget at the end of one window and again at the start of the next, passing
// twice the limit across the boundary.
//
// Keys are held in memory and never evicted, so every distinct key ever seen
// costs a map entry for the life of the process. Prefer key functions with a
// bounded range; keying on a rotating credential grows the map without limit.
type FixedWindow struct {
	mu        sync.Mutex
	reqLimit  int
	windowLen time.Duration
	windows   map[string]*window
}

// NewFixedWindow returns a [FixedWindow] granting reqLimit
// requests per key per windowLen.
//
// The returned limiter is one budget. Share it to pool routes under a single
// limit, or build separate ones to keep their limits independent.
func NewFixedWindow(
	reqLimit int,
	windowLen time.Duration,
) *FixedWindow {
	return &FixedWindow{
		reqLimit:  reqLimit,
		windowLen: windowLen,
		windows:   make(map[string]*window),
	}
}

// Limit counts one request against key and reports whether it may proceed. It
// is safe for concurrent use and never returns an error; the error result exists
// for limiters backed by a store that can fail.
func (fw *FixedWindow) Limit(ctx context.Context, key string) (LimitResult, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()
	win, ok := fw.windows[key]
	if !ok || !now.Before(win.end) {
		win = &window{
			end:   now.Add(fw.windowLen),
			count: 0,
		}
		fw.windows[key] = win
	}

	win.count++

	res := LimitResult{
		Allowed:    win.count <= fw.reqLimit,
		Limit:      fw.reqLimit,
		ResetAt:    win.end,
		RetryAfter: win.end.Sub(now),
	}

	if res.Allowed {
		res.Remaining = fw.reqLimit - win.count
	} else {
		res.Remaining = 0
	}

	return res, nil
}

type window struct {
	end   time.Time
	count int
}
