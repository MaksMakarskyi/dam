package keyfunc

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

var _ KeyFunc = IP

// IP keys a request by the client's IP address, giving every address its own
// budget. It reads the first entry of the X-Forwarded-For header, which by
// convention holds the originating client, and falls back to the address of
// the connection itself. The IPv6 loopback ::1 is normalised to 127.0.0.1 so
// that local traffic shares one bucket whichever stack it arrives on.
//
// IP returns an error when no valid address can be recovered from either
// source.
//
// Note that X-Forwarded-For is set by the caller and is trivially spoofed.
// Only use IP behind a proxy that overwrites the header on every request; if
// clients can reach the server directly, they can vary the header to get a
// fresh budget per request and bypass the limit entirely.
func IP(r *http.Request) (string, error) {
	ips := r.Header.Get("X-Forwarded-For")
	splitIps := strings.Split(ips, ",")
	clientIp := strings.TrimSpace(splitIps[0])

	if len(clientIp) > 0 {
		// Each proxy appends itself on the way through, so the list grows
		// left to right and the leftmost entry is the original client:
		//
		//	X-Forwarded-For: <client>, <proxy1>, …, <proxyN>
		//
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/X-Forwarded-For
		netIP := net.ParseIP(clientIp)
		if netIP != nil {
			return netIP.String(), nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	netIP := net.ParseIP(ip)
	if netIP != nil {
		ip := netIP.String()
		if ip == "::1" {
			return "127.0.0.1", nil
		}

		return ip, nil
	}

	return "", errors.New("IP not found")
}
