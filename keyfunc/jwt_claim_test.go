package keyfunc

import (
	"encoding/base64"
	"errors"
	"net/http/httptest"
	"testing"
)

func testJWT(payload string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims := base64.RawURLEncoding.EncodeToString([]byte(payload))

	return header + "." + claims + ".signature"
}

func TestJWTClaim(t *testing.T) {
	tests := map[string]struct {
		claim   string
		token   string
		wantKey string
		wantErr error
	}{
		"sub":               {claim: "sub", token: testJWT(`{"sub":"9f8b-user"}`), wantKey: "9f8b-user", wantErr: nil},
		"tenant_claim":      {claim: "org_id", token: testJWT(`{"sub":"9f8b-user","org_id":"acme"}`), wantKey: "acme", wantErr: nil},
		"mixed_value_types": {claim: "sub", token: testJWT(`{"sub":"9f8b-user","exp":1754624000,"meta":{"p":"email"}}`), wantKey: "9f8b-user", wantErr: nil},
		"claim_absent":      {claim: "org_id", token: testJWT(`{"sub":"9f8b-user"}`), wantKey: "", wantErr: ErrNoJWTClaim},
		"claim_empty":       {claim: "sub", token: testJWT(`{"sub":""}`), wantKey: "", wantErr: ErrNoJWTClaim},
		"claim_not_string":  {claim: "exp", token: testJWT(`{"exp":1754624000}`), wantKey: "", wantErr: ErrNoJWTClaim},
		"not_a_jwt":         {claim: "sub", token: "sk_live_abc", wantKey: "", wantErr: ErrNoJWTClaim},
		"payload_not_b64":   {claim: "sub", token: "header.!not-base64!.signature", wantKey: "", wantErr: ErrNoJWTClaim},
		"payload_not_json":  {claim: "sub", token: testJWT(`not json`), wantKey: "", wantErr: ErrNoJWTClaim},
		"no_header":         {claim: "sub", token: "", wantKey: "", wantErr: ErrNoJWTClaim},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)

			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}

			key, err := JWTClaim(tc.claim)(req)

			if tc.wantKey != key {
				t.Errorf("(key) expected: %q, got: %q", tc.wantKey, key)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("(error) expected: %v, got: %v", tc.wantErr, err)
			}
		})
	}
}
