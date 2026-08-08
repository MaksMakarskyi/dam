package keyfunc

import "net/http"

type KeyFunc func(r *http.Request) (string, error)
