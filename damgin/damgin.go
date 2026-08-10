// Package damgin adapts dam's rate limiting to the Gin framework.
//
// It is a module of its own, so programs on net/http or chi never pull Gin
// into their build:
//
//	go get github.com/MaksMakarskyi/dam/damgin
//
// Limiters and key functions come from the root dam package; only the handler
// signature differs. A rejected request is aborted, so no handler further down
// the chain runs, and the cause of a failure is attached with
// [gin.Context.Error] for whatever logging middleware is installed.
package damgin

import (
	"net/http"

	"github.com/MaksMakarskyi/dam"
	"github.com/gin-gonic/gin"
)

// Limit returns middleware that rate limits everything registered after it,
// keyed by keyFn:
//
//	router.Use(damgin.Limit(l, dam.KeyByIP))
//
// Gin middleware and handlers share one type, so the same value also guards a
// single route when passed ahead of its handler:
//
//	router.GET("/hi", damgin.Limit(l, dam.KeyByIP), sayHi)
//
// One limiter is one budget. Every handler the returned middleware guards draws
// from l, so a router-wide Use puts all routes under a shared limit; pass
// separate limiters to give them separate limits. The same holds when stacking
// layers — reusing one limiter at both the router and the route level charges
// each request against it twice.
//
// A nil l or keyFn is replaced as described on [LimitFunc]. The substitution
// happens once, here, so a defaulted limiter is shared across the guarded
// handlers exactly like an explicit one.
func Limit(l dam.Limiter, keyFn dam.KeyFunc) gin.HandlerFunc {
	l, keyFn = dam.OrDefaults(l, keyFn)

	return func(c *gin.Context) {
		if allow(c, l, keyFn) {
			c.Next()
		}
	}
}

// LimitFunc wraps next and returns the rate limited [gin.HandlerFunc]:
//
//	router.GET("/hi", damgin.LimitFunc(l, dam.KeyByIP, sayHi))
//
// It is the explicit form of [Limit] for a single handler; prefer whichever
// reads better, as Gin accepts both.
//
// Every response carries RateLimit-Limit, RateLimit-Remaining and
// RateLimit-Reset. Requests over the limit are answered with 429 and a
// Retry-After header, the chain is aborted, and next is not called.
//
// A nil l is replaced by a fixed window of [dam.DefaultLimit] requests per
// [dam.DefaultWindow], built fresh on every call, so handlers wrapped by
// separate calls get separate budgets. A nil keyFn is replaced by
// [dam.KeyGlobal], which counts every request into one bucket: that is a
// service-wide cap rather than a per-client one, so a single busy client can
// exhaust it for everyone. Pass [dam.KeyByIP] or another key function to give
// clients their own budgets.
//
// A key function that fails yields a 500. [dam.KeyByAPIKey] and
// [dam.KeyByJWTClaim] fail on unauthenticated requests, where 401 is the
// correct answer, so place them behind authentication middleware.
func LimitFunc(
	l dam.Limiter,
	keyFn dam.KeyFunc,
	next gin.HandlerFunc,
) gin.HandlerFunc {
	l, keyFn = dam.OrDefaults(l, keyFn)

	return func(c *gin.Context) {
		if allow(c, l, keyFn) {
			next(c)
		}
	}
}

// allow charges one request against l and reports whether it may proceed. It
// writes the RateLimit headers either way, and on a rejection or a failure it
// aborts c so nothing further in the chain runs.
func allow(c *gin.Context, l dam.Limiter, keyFn dam.KeyFunc) bool {
	key, err := keyFn(c.Request)
	if err != nil {
		_ = c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get a rate limiting key for request",
		})
		return false
	}

	result, err := l.Limit(c.Request.Context(), key)
	if err != nil {
		_ = c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to limit the request",
		})
		return false
	}

	dam.SetRateLimitHeaders(c.Writer, result)
	if !result.Allowed {
		dam.SetRetryAfter(c.Writer, result)
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": "rate limit exceeded",
		})
		return false
	}

	return true
}
