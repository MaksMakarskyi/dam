package dam

import (
	"net/http"
	"time"
)

// Defaults substituted for a nil limiter, as described on [LimitFunc].
const (
	DefaultLimit  = 10
	DefaultWindow = time.Second
)

// Limit returns middleware that rate limits the handlers it wraps, keyed by
// keyFn. The returned value has the shape every net/http router composes with,
// so it also serves as chi middleware:
//
//	r.Use(dam.Limit(l, dam.KeyByIP))
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
func Limit(l Limiter, keyFn KeyFunc) func(http.Handler) http.Handler {
	l, keyFn = OrDefaults(l, keyFn)

	return func(h http.Handler) http.Handler {
		return LimitHandler(l, keyFn, h)
	}
}

// LimitHandler wraps next and returns the rate limited [http.Handler]. It is
// the single-handler form of [Limit]; next must not be nil. A nil l or keyFn is
// replaced as described on [LimitFunc].
func LimitHandler(l Limiter, keyFn KeyFunc, next http.Handler) http.Handler {
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
// calls get separate budgets. A nil keyFn is replaced by [KeyGlobal], which
// counts every request into one bucket: that is a service-wide cap rather than
// a per-client one, so a single busy client can exhaust it for everyone. Pass
// [KeyByIP] or another key function to give clients their own budgets.
//
// A key function that fails is answered with 500. [KeyByAPIKey] and
// [KeyByJWTClaim] fail on unauthenticated requests, where 401 is the correct
// answer, so place them behind authentication middleware.
func LimitFunc(
	l Limiter,
	keyFn KeyFunc,
	next func(w http.ResponseWriter, r *http.Request),
) func(w http.ResponseWriter, r *http.Request) {
	l, keyFn = OrDefaults(l, keyFn)

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

		SetRateLimitHeaders(w, result)
		if !result.Allowed {
			SetRetryAfter(w, result)
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

// OrDefaults returns l and keyFn unchanged, substituting the package defaults
// for whichever of them is nil: a fresh [FixedWindow] of [DefaultLimit] per
// [DefaultWindow] for the limiter, [KeyGlobal] for the key function.
//
// Run it once per wrap, never per request. A limiter it builds is shared by
// every request reaching the wrapped handler, exactly like an explicit one;
// calling it per request would mint a new budget each time and limit nothing.
//
// It is exported for adapters targeting frameworks this module does not ship,
// so their nil handling matches the middleware here. [Limit], [LimitHandler]
// and [LimitFunc] already call it.
func OrDefaults(l Limiter, keyFn KeyFunc) (Limiter, KeyFunc) {
	if l == nil {
		l = NewFixedWindow(DefaultLimit, DefaultWindow)
	}
	if keyFn == nil {
		keyFn = KeyGlobal
	}

	return l, keyFn
}
