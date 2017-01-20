package types

// Route is the stored route for a proxied service
type Route struct {
	ID         string
	TargetPort int
	ProxiedURL string `json:",omitempty"`
}
