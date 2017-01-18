package types

import (
	"net/http/httputil"
	"regexp"
)

// Route
type Route struct {
	ID            string
	TargetURL     string
	IncomingMatch string                 `json:",omitempty"`
	Regexp        *regexp.Regexp         `json:"-"`
	Proxy         *httputil.ReverseProxy `json:"-"`
}
