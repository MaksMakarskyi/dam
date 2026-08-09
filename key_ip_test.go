package dam

import (
	"errors"
	"net/http/httptest"
	"testing"
)

func TestIP(t *testing.T) {
	tests := map[string]struct {
		xForwardedFor string
		remoteAddr    string
		wantKey       string
		wantErr       error
	}{
		"xff_single":                           {xForwardedFor: "203.0.113.5", remoteAddr: "192.0.2.1:1234", wantKey: "203.0.113.5", wantErr: nil},
		"xff_chain":                            {xForwardedFor: "203.0.113.5, 70.41.3.18, 150.172.238.178", remoteAddr: "192.0.2.1:1234", wantKey: "203.0.113.5", wantErr: nil},
		"xff_surrounding_spaces":               {xForwardedFor: "   203.0.113.5   , 70.41.3.18", remoteAddr: "192.0.2.1:1234", wantKey: "203.0.113.5", wantErr: nil},
		"xff_ipv6_canonicalised":               {xForwardedFor: "2001:0DB8::0001", remoteAddr: "192.0.2.1:1234", wantKey: "2001:db8::1", wantErr: nil},
		"xff_ipv4_mapped_ipv6":                 {xForwardedFor: "::ffff:192.0.2.1", remoteAddr: "198.51.100.7:1234", wantKey: "192.0.2.1", wantErr: nil},
		"xff_invalid_falls_back":               {xForwardedFor: "not-an-ip", remoteAddr: "192.0.2.1:1234", wantKey: "192.0.2.1", wantErr: nil},
		"xff_blank_falls_back":                 {xForwardedFor: "   ", remoteAddr: "192.0.2.1:1234", wantKey: "192.0.2.1", wantErr: nil},
		"remote_addr_ipv4":                     {xForwardedFor: "", remoteAddr: "192.0.2.1:1234", wantKey: "192.0.2.1", wantErr: nil},
		"remote_addr_ipv6":                     {xForwardedFor: "", remoteAddr: "[2001:db8::1]:443", wantKey: "2001:db8::1", wantErr: nil},
		"remote_addr_loopback":                 {xForwardedFor: "", remoteAddr: "127.0.0.1:8080", wantKey: "127.0.0.1", wantErr: nil},
		"remote_addr_loopback_ipv6_normalised": {xForwardedFor: "", remoteAddr: "[::1]:443", wantKey: "127.0.0.1", wantErr: nil},
		"xff_loopback_ipv6_normalised":         {xForwardedFor: "::1", remoteAddr: "192.0.2.1:1234", wantKey: "127.0.0.1", wantErr: nil},
		"remote_addr_hostname":                 {xForwardedFor: "", remoteAddr: "example.com:80", wantKey: "", wantErr: ErrNoIP},
		"remote_addr_missing_port":             {xForwardedFor: "", remoteAddr: "192.0.2.1", wantKey: "", wantErr: ErrNoIP},
		"remote_addr_empty":                    {xForwardedFor: "", remoteAddr: "", wantKey: "", wantErr: ErrNoIP},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tc.remoteAddr

			if tc.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tc.xForwardedFor)
			}

			key, err := KeyByIP(req)

			if tc.wantKey != key {
				t.Errorf("(key) expected: %q, got: %q", tc.wantKey, key)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("(error) expected: %v, got: %v", tc.wantErr, err)
			}
		})
	}
}
