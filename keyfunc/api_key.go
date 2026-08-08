package keyfunc

import (
	"errors"
	"net/http"
	"strings"
)

var ErrNoApiKey = errors.New("failed to retrieve the API key")

var _ KeyFunc = ApiKey

// ApiKey keys a request by the bearer token it presents, giving every API key
// its own budget.
//
// Only the standard bearer scheme is accepted: "Authorization: Bearer
// sk_live_abc" yields "sk_live_abc". The scheme name is matched
// case-insensitively, as RFC 9110 requires. Any other scheme, and a bare key
// sent without one, is rejected rather than keyed on.
//
// ApiKey returns [ErrNoApiKey] for anything it cannot key: a missing or
// blank header, an unsupported scheme, or Bearer with no token after it.
// Callers usually want to answer that with 401, since the request was never
// authenticated.
//
// The key returned is the raw secret, retained as a map key inside the
// limiter; hash it first if bearer tokens must not sit in process memory.
func ApiKey(r *http.Request) (string, error) {
	return getApiKey(r)
}

func getApiKey(r *http.Request) (string, error) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth == "" {
		return "", ErrNoApiKey
	}

	scheme, token, _ := strings.Cut(auth, " ")
	if !strings.EqualFold(scheme, "Bearer") {
		return "", ErrNoApiKey
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", ErrNoApiKey
	}

	return token, nil
}
