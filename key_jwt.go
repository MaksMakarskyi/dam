package dam

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var ErrNoJWTClaim = errors.New("failed to retrieve the JWT claim")

var _ KeyFunc = KeyByJWTClaim("sub")

// KeyByJWTClaim returns a [KeyFunc] that keys a request by a named string claim of
// the JWT it presents, giving every distinct claim value its own budget.
//
// The token is taken from the Authorization header as a bearer credential and
// its payload is base64url-decoded to read the claim. Worthwhile choices are
// "sub" for the end user, which is the only identity claim RFC 7519 registers,
// and a tenant claim such as "org_id" when a whole customer should share one
// budget.
//
// The returned KeyFunc reports [ErrNoJWTClaim] when the request carries no
// bearer token, when that token is not a well-formed JWT, or when the claim is
// absent, empty, or not a string.
//
// KeyByJWTClaim verifies nothing. The signature is never checked, so the value it
// returns is whatever the client typed into the header. A client can rotate
// invented claim values to shed its own limit, or send another user's
// identifier to drain that user's budget. Use it only behind middleware that
// has already authenticated the request.
func KeyByJWTClaim(name string) KeyFunc {
	return func(r *http.Request) (string, error) {
		token, err := getAPIKey(r)
		if err != nil {
			return "", fmt.Errorf("%w %q: %w", ErrNoJWTClaim, name, err)
		}

		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			return "", fmt.Errorf("%w %q: not a JWT", ErrNoJWTClaim, name)
		}

		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return "", fmt.Errorf("%w %q: payload is not valid base64url", ErrNoJWTClaim, name)
		}

		var payload map[string]json.RawMessage
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			return "", fmt.Errorf("%w %q: payload is not valid JSON", ErrNoJWTClaim, name)
		}

		raw, ok := payload[name]
		if !ok {
			return "", fmt.Errorf("%w %q: claim is absent", ErrNoJWTClaim, name)
		}

		var claim string
		if err := json.Unmarshal(raw, &claim); err != nil {
			return "", fmt.Errorf("%w %q: claim is not a string", ErrNoJWTClaim, name)
		}
		if claim == "" {
			return "", fmt.Errorf("%w %q: claim is empty", ErrNoJWTClaim, name)
		}

		return claim, nil
	}
}
