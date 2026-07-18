package webclaw

import "context"

// XAudienceDirection selects which side of an account's graph to export.
type XAudienceDirection string

const (
	// XAudienceFollowers exports the accounts that follow the target (default).
	XAudienceFollowers XAudienceDirection = "followers"
	// XAudienceFollowing exports the accounts the target follows.
	XAudienceFollowing XAudienceDirection = "following"
)

// XAudienceRequest configures an X (Twitter) audience export. Provide either
// Handle or UserID -- Handle is resolved once (unbilled) to a numeric id; pass
// the returned UserID back on later pages to skip re-resolving.
//
// This is a paid-only, metered feature: the API returns 403 for free or
// lapsed accounts, and each page fetched costs one credit.
type XAudienceRequest struct {
	// Handle is an @handle (the leading @ is optional). Resolved once, unbilled.
	// Provide Handle or UserID.
	Handle string `json:"handle,omitempty"`
	// UserID is a pre-resolved numeric id. Passing it back on later pages skips
	// re-resolving the handle.
	UserID string `json:"user_id,omitempty"`
	// Direction is "followers" (default) or "following".
	Direction XAudienceDirection `json:"direction,omitempty"`
	// Cursor is the opaque NextCursor from a previous response. Omit on the
	// first page.
	Cursor string `json:"cursor,omitempty"`
	// MaxPages caps how many pages this call walks. Defaults to 2, clamped to
	// 1..10. Each page is roughly 1-2k users.
	MaxPages int `json:"max_pages,omitempty"`
}

// XAudienceUser is a single account in an exported audience.
type XAudienceUser struct {
	ID          string `json:"id"`
	ScreenName  string `json:"screen_name"`
	Name        string `json:"name"`
	Followers   int    `json:"followers"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// XAudienceResponse is one metered slice of an exported audience. When
// NextCursor is nil the audience has been fully walked; otherwise call
// ExportXAudience again, passing back UserID and NextCursor, until it is nil.
type XAudienceResponse struct {
	UserID    string             `json:"user_id"`
	Direction XAudienceDirection `json:"direction"`
	Count     int                `json:"count"`
	Users     []XAudienceUser    `json:"users"`
	// NextCursor is the cursor to resume from. It is nil once the audience is
	// fully walked; pass its value back as XAudienceRequest.Cursor otherwise.
	NextCursor     *string `json:"next_cursor"`
	PagesFetched   int     `json:"pages_fetched"`
	CreditsCharged int     `json:"credits_charged"`
}

// ExportXAudience exports the followers or following of an X account, one
// cursor-paginated and metered slice per call (one credit per page fetched).
//
// To walk a full audience, call repeatedly, passing back the returned UserID
// and NextCursor, until NextCursor is nil:
//
//	req := &webclaw.XAudienceRequest{Handle: "jack", Direction: webclaw.XAudienceFollowers}
//	for {
//		page, err := client.ExportXAudience(ctx, req)
//		if err != nil {
//			log.Fatal(err)
//		}
//		for _, u := range page.Users {
//			fmt.Println(u.ScreenName)
//		}
//		if page.NextCursor == nil {
//			break
//		}
//		req.UserID = page.UserID
//		req.Cursor = *page.NextCursor
//	}
func (c *Client) ExportXAudience(ctx context.Context, req *XAudienceRequest) (*XAudienceResponse, error) {
	var resp XAudienceResponse
	if err := c.do(ctx, "POST", "/v1/x/audience", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
