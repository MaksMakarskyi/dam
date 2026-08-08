package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MaksMakarskyi/dam/middleware"
)

func MockKeyFunc(r *http.Request) (string, error) {
	return "id-123", nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

func TestFixedWindow(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	tests := map[string]struct {
		limit     int
		windowLen time.Duration
		keyFn     func(r *http.Request) (string, error)
		intervals []time.Duration
		responses []int
	}{
		"all_pass": {
			limit:     1,
			windowLen: time.Millisecond,
			keyFn:     MockKeyFunc,
			intervals: []time.Duration{2 * time.Millisecond, 3 * time.Millisecond},
			responses: []int{http.StatusOK, http.StatusOK, http.StatusOK},
		},
		"tight_intervals": {
			limit:     1,
			windowLen: time.Millisecond,
			keyFn:     MockKeyFunc,
			intervals: []time.Duration{time.Millisecond, time.Millisecond},
			responses: []int{http.StatusOK, http.StatusOK, http.StatusOK},
		},
		"fail_between_ok": {
			limit:     1,
			windowLen: 2 * time.Millisecond,
			keyFn:     MockKeyFunc,
			intervals: []time.Duration{time.Millisecond, 2 * time.Millisecond},
			responses: []int{http.StatusOK, http.StatusTooManyRequests, http.StatusOK},
		},
		"exceed_limit_three": {
			limit:     3,
			windowLen: 3 * time.Millisecond,
			keyFn:     MockKeyFunc,
			intervals: []time.Duration{time.Millisecond, time.Millisecond, time.Microsecond},
			responses: []int{http.StatusOK, http.StatusOK, http.StatusOK, http.StatusTooManyRequests},
		},
		"exceed_limit_five": {
			limit:     5,
			windowLen: 5 * time.Millisecond,
			keyFn:     MockKeyFunc,
			intervals: []time.Duration{time.Millisecond, time.Millisecond, time.Microsecond, time.Millisecond, time.Microsecond},
			responses: []int{http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK, http.StatusTooManyRequests},
		},
	}

	for name, tc := range tests {
		// Guard against a malformed table entry: intervals are the gaps
		// between consecutive requests, so there is one fewer of them.
		if len(tc.intervals) != len(tc.responses)-1 {
			t.Fatalf(
				"malformed test case %q: %d responses need %d intervals between them, got %d",
				name, len(tc.responses), len(tc.responses)-1, len(tc.intervals),
			)
		}

		h := middleware.LimitFixedWindow(tc.limit, tc.windowLen, tc.keyFn)(http.HandlerFunc(Handler))
		t.Run(name, func(t *testing.T) {
			for i, expected := range tc.responses {
				w := httptest.NewRecorder()
				h.ServeHTTP(w, req)

				res := w.Result()
				defer res.Body.Close()

				if res.StatusCode != expected {
					t.Errorf("(StatusCode) Expected: %d, Got: %d", expected, res.StatusCode)
				}

				if i < len(tc.responses)-1 {
					time.Sleep(tc.intervals[i])
				}
			}
		})
	}
}
