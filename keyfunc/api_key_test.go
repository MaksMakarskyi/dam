package keyfunc

import (
	"errors"
	"net/http/httptest"
	"testing"
)

func TestApiKey(t *testing.T) {
	tests := map[string]struct {
		header  string
		wantKey string
		wantErr error
	}{
		"standard":           {header: "Bearer sk_live_abc", wantKey: "sk_live_abc", wantErr: nil},
		"lowercase":          {header: "bearer sk_live_abc", wantKey: "sk_live_abc", wantErr: nil},
		"uppercase":          {header: "BEARER sk_live_abc", wantKey: "sk_live_abc", wantErr: nil},
		"surrounding_spaces": {header: "   Bearer      sk_live_abc    ", wantKey: "sk_live_abc", wantErr: nil},
		"basic_scheme":       {header: "Basic dXNlcjpwdw==", wantKey: "", wantErr: ErrNoApiKey},
		"digest_scheme":      {header: "Digest xyz", wantKey: "", wantErr: ErrNoApiKey},
		"no_scheme":          {header: "sk_live_abc", wantKey: "", wantErr: ErrNoApiKey},
		"no_key":             {header: "Bearer", wantKey: "", wantErr: ErrNoApiKey},
		"empty_key":          {header: "Bearer    ", wantKey: "", wantErr: ErrNoApiKey},
		"empty_header":       {header: "    ", wantKey: "", wantErr: ErrNoApiKey},
		"no_header":          {header: "", wantKey: "", wantErr: ErrNoApiKey},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)

			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}

			key, err := ApiKey(req)

			if tc.wantKey != key {
				t.Errorf("(key) expected: %q, got: %q", tc.wantKey, key)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("(error) expected: %v, got: %v", tc.wantErr, err)
			}
		})
	}
}
