package dam

import (
	"errors"
	"net/http"
	"strings"
)

var ErrNoAPIKey = errors.New("failed to retrieve the API key")

var _ KeyFunc = KeyByAPIKey

// KeyByAPIKey keys a request by the bearer token it presents, giving every API key
// its own budget.
//
// Only the standard bearer scheme is accepted: "Authorization: Bearer
// sk_live_abc" yields "sk_live_abc". The scheme name is matched
// case-insensitively, as RFC 9110 requires. Any other scheme, and a bare key
// sent without one, is rejected rather than keyed on.
//
// KeyByAPIKey returns [ErrNoAPIKey] for anything it cannot key: a missing or
// blank header, an unsupported scheme, or Bearer with no token after it.
// Callers usually want to answer that with 401, since the request was never
// authenticated.
//
// The key returned is the raw secret, retained as a map key inside the
// limiter; hash it first if bearer tokens must not sit in process memory.
func KeyByAPIKey(r *http.Request) (string, error) {
	return getAPIKey(r)
}

// getAPIKey extracts the bearer token from the Authorization header. It backs
// both [KeyByAPIKey] and [KeyByJWTClaim], which differ only in what they do
// with the token afterwards.
func getAPIKey(r *http.Request) (string, error) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth == "" {
		return "", ErrNoAPIKey
	}

	scheme, token, _ := strings.Cut(auth, " ")
	if !strings.EqualFold(scheme, "Bearer") {
		return "", ErrNoAPIKey
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", ErrNoAPIKey
	}

	return token, nil
}
