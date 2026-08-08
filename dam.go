package dam

import (
	"net/http"
	"time"

	"github.com/MaksMakarskyi/dam/keyfunc"
	"github.com/MaksMakarskyi/dam/limiter"
)

const (
	DefaultLimit  = 10
	DefaultWindow = time.Second
)

// Limit returns middleware that rate limits the handlers it wraps, keyed by
// keyFn. The returned value has the shape every net/http router composes with,
// so it also serves as chi middleware:
//
//	r.Use(dam.Limit(l, keyfunc.IP))
//
// One limiter is one budget. Every handler the returned middleware wraps draws
// from l, so wrapping three routes gives those three routes a shared limit;
// pass separate limiters to give them separate limits. The same holds when
// stacking layers — reusing one limiter at both the router and the route level
// charges each request twice.
//
// A nil l or keyFn is replaced as described on [LimitFunc]. The substitution
// happens once, here, so a defaulted limiter is shared across the wrapped
// handlers exactly like an explicit one.
func Limit(l limiter.Limiter, keyFn keyfunc.KeyFunc) func(http.Handler) http.Handler {
	l, keyFn = withDefaults(l, keyFn)

	return func(h http.Handler) http.Handler {
		return LimitHandler(l, keyFn, h)
	}
}

// LimitHandler wraps next and returns the rate limited [http.Handler]. It is
// the single-handler form of [Limit]; next must not be nil. A nil l or keyFn is
// replaced as described on [LimitFunc].
func LimitHandler(l limiter.Limiter, keyFn keyfunc.KeyFunc, next http.Handler) http.Handler {
	return http.HandlerFunc(LimitFunc(l, keyFn, next.ServeHTTP))
}

// LimitFunc wraps next and returns the rate limited handler function, ready for
// [http.ServeMux.HandleFunc].
//
// Every response carries RateLimit-Limit, RateLimit-Remaining and
// RateLimit-Reset. Requests over the limit are answered with 429 and a
// Retry-After header, and next is not called.
//
// A nil l is replaced by a fixed window of [DefaultLimit] requests per
// [DefaultWindow], built fresh on every call, so handlers wrapped by separate
// calls get separate budgets. A nil keyFn is replaced by [keyfunc.Common],
// which counts every request into one bucket: that is a service-wide cap rather
// than a per-client one, so a single busy client can exhaust it for everyone.
// Pass [keyfunc.IP] or another key function to give clients their own budgets.
//
// A key function that fails is answered with 500. [keyfunc.ApiKey] and
// [keyfunc.JWTClaim] fail on unauthenticated requests, where 401 is the correct
// answer, so place them behind authentication middleware.
func LimitFunc(
	l limiter.Limiter,
	keyFn keyfunc.KeyFunc,
	next func(w http.ResponseWriter, r *http.Request),
) func(w http.ResponseWriter, r *http.Request) {
	l, keyFn = withDefaults(l, keyFn)

	return func(w http.ResponseWriter, r *http.Request) {
		key, err := keyFn(r)
		if err != nil {
			http.Error(w, "failed to get a rate limiting key for request", http.StatusInternalServerError)
			return
		}

		result, err := l.Limit(r.Context(), key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		setHeaders(w, result)
		if !result.Allowed {
			setLimitedHeaders(w, result)
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

func withDefaults(l limiter.Limiter, keyFn keyfunc.KeyFunc) (limiter.Limiter, keyfunc.KeyFunc) {
	if l == nil {
		l = limiter.NewFixedWindowLimiter(DefaultLimit, DefaultWindow)
	}
	if keyFn == nil {
		keyFn = keyfunc.Common
	}

	return l, keyFn
}
