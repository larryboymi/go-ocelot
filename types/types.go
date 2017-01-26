package types

// Route is the stored route for a proxied service
type Route struct {
	ID          string `json:"id"`
	TargetPort  int    `json:"targetPort"`
	Description string `json:"description,omitempty"`
	ProxiedURL  string `json:"proxiedURL,omitempty"`
}
