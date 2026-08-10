// Package damecho adapts dam's rate limiting to the Echo framework.
//
// It is a module of its own, so programs on net/http or chi never pull Echo
// into their build:
//
//	go get github.com/MaksMakarskyi/dam/damecho
//
// Limiters and key functions come from the root dam package; only the handler
// signature differs. Rejections are returned as [echo.HTTPError] rather than
// written straight to the response, so Echo's HTTPErrorHandler renders them in
// whatever format the application already uses.
package damecho

import (
	"net/http"

	"github.com/MaksMakarskyi/dam"
	"github.com/labstack/echo/v5"
)

// Limit returns middleware that rate limits the handlers it wraps, keyed by
// keyFn:
//
//	e.Use(damecho.Limit(l, dam.KeyByIP))
//
// One limiter is one budget. Every handler the returned middleware wraps draws
// from l, so a router-wide Use puts all routes under a shared limit; pass
// separate limiters to give them separate limits. The same holds when stacking
// layers — reusing one limiter at both the router and the route level charges
// each request against it twice.
//
// A nil l or keyFn is replaced as described on [LimitFunc]. The substitution
// happens once, here, so a defaulted limiter is shared across the wrapped
// handlers exactly like an explicit one.
func Limit(l dam.Limiter, keyFn dam.KeyFunc) echo.MiddlewareFunc {
	l, keyFn = dam.OrDefaults(l, keyFn)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return LimitFunc(l, keyFn, next)
	}
}

// LimitFunc wraps next and returns the rate limited [echo.HandlerFunc], ready
// to register on a single route:
//
//	e.GET("/hi", damecho.LimitFunc(l, dam.KeyByIP, sayHi))
//
// Every response carries RateLimit-Limit, RateLimit-Remaining and
// RateLimit-Reset. Requests over the limit are rejected with an
// [echo.HTTPError] of status 429 plus a Retry-After header, and next is not
// called.
//
// A nil l is replaced by a fixed window of [dam.DefaultLimit] requests per
// [dam.DefaultWindow], built fresh on every call, so handlers wrapped by
// separate calls get separate budgets. A nil keyFn is replaced by
// [dam.KeyGlobal], which counts every request into one bucket: that is a
// service-wide cap rather than a per-client one, so a single busy client can
// exhaust it for everyone. Pass [dam.KeyByIP] or another key function to give
// clients their own budgets.
//
// A key function that fails yields a 500 [echo.HTTPError] wrapping its error.
// [dam.KeyByAPIKey] and [dam.KeyByJWTClaim] fail on unauthenticated requests,
// where 401 is the correct answer, so place them behind authentication
// middleware.
func LimitFunc(
	l dam.Limiter,
	keyFn dam.KeyFunc,
	next echo.HandlerFunc,
) echo.HandlerFunc {
	l, keyFn = dam.OrDefaults(l, keyFn)

	return func(c *echo.Context) error {
		req := c.Request()

		key, err := keyFn(req)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				"failed to get a rate limiting key for request",
			).Wrap(err)
		}

		result, err := l.Limit(req.Context(), key)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				"failed to limit the request",
			).Wrap(err)
		}

		res := c.Response()

		dam.SetRateLimitHeaders(res, result)
		if !result.Allowed {
			dam.SetRetryAfter(res, result)
			return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
		}

		return next(c)
	}
}
