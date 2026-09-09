package webclaw

import "context"

// DiffRequest configures a page diff request.
type DiffRequest struct {
	URL string `json:"url"`
	// Previous must be a complete extraction with metadata and content.
	// Omit to use this caller's most recent cached extraction.
	Previous map[string]interface{} `json:"previous,omitempty"`
}

type DiffMetadataChange struct {
	Field string  `json:"field"`
	Old   *string `json:"old"`
	New   *string `json:"new"`
}

type DiffLink struct {
	Href string `json:"href"`
	Text string `json:"text"`
}

// DiffResponse contains the detected changes.
type DiffResponse struct {
	Status          string               `json:"status"` // Same, Changed, or New
	TextDiff        *string              `json:"text_diff"`
	MetadataChanges []DiffMetadataChange `json:"metadata_changes"`
	LinksAdded      []DiffLink           `json:"links_added"`
	LinksRemoved    []DiffLink           `json:"links_removed"`
	WordCountDelta  int64                `json:"word_count_delta"`
}

// Diff compares the current state of a page against a previous snapshot.
func (c *Client) Diff(ctx context.Context, req *DiffRequest) (*DiffResponse, error) {
	var resp DiffResponse
	if err := c.do(ctx, "POST", "/v1/diff", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
