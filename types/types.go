package types

// Route
type Route struct {
	ID         string
	TargetPort int
	ProxiedURL string `json:",omitempty"`
}
