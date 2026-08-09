package dam

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

var ErrNoIP = errors.New("failed to retrieve the IP address")

var _ KeyFunc = KeyByIP

// KeyByIP keys a request by the client's IP address, giving every address its own
// budget. It reads the first entry of the X-Forwarded-For header, which by
// convention holds the originating client, and falls back to the address of
// the connection itself. The IPv6 loopback ::1 is normalised to 127.0.0.1 so
// that local traffic shares one bucket whichever stack it arrives on.
//
// KeyByIP returns an error when no valid address can be recovered from either
// source.
//
// Note that X-Forwarded-For is set by the caller and is trivially spoofed.
// Only use KeyByIP behind a proxy that overwrites the header on every request; if
// clients can reach the server directly, they can vary the header to get a
// fresh budget per request and bypass the limit entirely.
func KeyByIP(r *http.Request) (string, error) {
	ips := r.Header.Get("X-Forwarded-For")
	splitIps := strings.Split(ips, ",")

	// Each proxy appends itself on the way through, so the list grows
	// left to right and the leftmost entry is the original client:
	//
	//	X-Forwarded-For: <client>, <proxy1>, …, <proxyN>
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/X-Forwarded-For
	clientIp := strings.TrimSpace(splitIps[0])
	if clientIp != "" {
		netIP := net.ParseIP(clientIp)
		if netIP != nil {
			ip := netIP.String()
			return normalize(ip), nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNoIP, err)
	}

	netIP := net.ParseIP(ip)
	if netIP != nil {
		ip := netIP.String()
		return normalize(ip), nil
	}

	return "", ErrNoIP
}

func normalize(ip string) string {
	if ip == "::1" {
		return "127.0.0.1"
	}

	return ip
}
