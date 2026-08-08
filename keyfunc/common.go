package keyfunc

import "net/http"

var _ KeyFunc = Common

// Common returns one fixed key for every request, placing all callers in a
// single shared bucket. Use it to cap the total throughput of an endpoint
// rather than the throughput of any individual client.
//
// The key is an arbitrary constant.
//
// Common never returns an error.
func Common(r *http.Request) (string, error) {
	return "dam_RjtAvBPDGzgBMnyUZE0!Kc5A$.Xzdfu8", nil
}
