package types

import "net/http/httputil"

// Route
type Route struct {
	ID        string
	TargetURL string
	Proxy     *httputil.ReverseProxy `json:"-"`
}
