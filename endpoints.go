package webclaw

import "context"

// EndpointKind classifies a discovered API endpoint.
type EndpointKind string

const (
	// EndpointKindRelativePath is a path-only reference (e.g. "/api/users").
	EndpointKindRelativePath EndpointKind = "relative_path"
	// EndpointKindAbsoluteURL is a fully-qualified HTTP(S) URL.
	EndpointKindAbsoluteURL EndpointKind = "absolute_url"
	// EndpointKindGraphQL is a GraphQL endpoint.
	EndpointKindGraphQL EndpointKind = "graph_ql"
	// EndpointKindWebSocket is a WebSocket endpoint (ws:// or wss://).
	EndpointKindWebSocket EndpointKind = "web_socket"
)

// EndpointsRequest configures an API endpoint discovery scan. It parses
// inline JavaScript and linked <script src> bundles to surface endpoints
// that a sitemap-based map cannot see.
type EndpointsRequest struct {
	URL string `json:"url"`
	// IncludeThirdParty also reports endpoints on hosts other than the
	// target. Defaults to false (first-party only).
	IncludeThirdParty bool `json:"include_third_party,omitempty"`
	// MaxBundles caps how many external script bundles are fetched and
	// scanned. Defaults to 20 server-side; 20 is also the maximum.
	MaxBundles int `json:"max_bundles,omitempty"`
}

// DiscoveredEndpoint is a single API endpoint found in the page's
// JavaScript.
type DiscoveredEndpoint struct {
	Value      string       `json:"value"`
	Kind       EndpointKind `json:"kind"`
	FirstParty bool         `json:"first_party"`
	Source     string       `json:"source"`
}

// EndpointsResponse contains the endpoints discovered for a URL.
type EndpointsResponse struct {
	URL            string               `json:"url"`
	BundlesScanned int                  `json:"bundles_scanned"`
	EndpointCount  int                  `json:"endpoint_count"`
	Endpoints      []DiscoveredEndpoint `json:"endpoints"`
	Hosts          []string             `json:"hosts"`
	Truncated      bool                 `json:"truncated"`
}

// Endpoints discovers API endpoints embedded in a page's inline JavaScript
// and linked script bundles -- references a sitemap-based Map cannot reach.
func (c *Client) Endpoints(ctx context.Context, req *EndpointsRequest) (*EndpointsResponse, error) {
	var resp EndpointsResponse
	if err := c.do(ctx, "POST", "/v1/endpoints", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
