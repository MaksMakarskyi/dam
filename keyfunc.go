package dam

import "net/http"

// KeyFunc derives the budget a request is counted against. Requests yielding
// equal keys share a budget; requests yielding different keys are independent.
//
// Returning an error rejects the request rather than keying it, which the
// middleware answers with 500. See [KeyByIP], [KeyByAPIKey], [KeyByJWTClaim]
// and [KeyGlobal] for the key functions this package provides.
type KeyFunc func(r *http.Request) (string, error)

var _ KeyFunc = KeyGlobal

// KeyGlobal returns one fixed key for every request, placing all callers in a
// single shared bucket. Use it to cap the total throughput of an endpoint
// rather than the throughput of any individual client.
//
// The key is an arbitrary constant.
//
// KeyGlobal never returns an error.
func KeyGlobal(r *http.Request) (string, error) {
	return "dam_RjtAvBPDGzgBMnyUZE0!Kc5A$.Xzdfu8", nil
}
