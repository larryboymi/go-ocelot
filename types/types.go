package types

import "net/http/httputil"

// Route
type Route struct {
	ID         string
	TargetPort int
	ProxiedURL string                 `json:",omitempty"`
	Proxy      *httputil.ReverseProxy `json:"-"`
}
