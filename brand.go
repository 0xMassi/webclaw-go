package webclaw

import (
	"context"
	"encoding/json"
	"fmt"
)

// Brand extracts brand identity information from a URL.
// The response is a flexible JSON object since the shape depends on the site.
func (c *Client) Brand(ctx context.Context, req *BrandRequest) (*BrandResponse, error) {
	var raw json.RawMessage
	if err := c.do(ctx, "POST", "/v1/brand", req, &raw); err != nil {
		return nil, err
	}
	return &BrandResponse{Data: raw}, nil
}

// Decode unmarshals the brand data into the provided struct.
func (r *BrandResponse) Decode(dst any) error {
	if r.Data == nil {
		return fmt.Errorf("webclaw: brand response has no data")
	}
	return json.Unmarshal(r.Data, dst)
}
