package webclaw

import "context"

// LeadRequest configures a lead enrichment lookup for a company URL.
type LeadRequest struct {
	URL string `json:"url"`
	// NoCache bypasses the cache and forces a fresh enrichment.
	NoCache bool `json:"no_cache,omitempty"`
}

// LeadSocials holds the social profile URLs discovered for a company.
type LeadSocials struct {
	LinkedIn string `json:"linkedin,omitempty"`
	X        string `json:"x,omitempty"`
	GitHub   string `json:"github,omitempty"`
}

// LeadPricing is a single pricing plan discovered on the company's site.
type LeadPricing struct {
	Plan  string `json:"plan"`
	Price string `json:"price"`
}

// LeadEmail is a categorized contact email discovered for the company
// (e.g. Type "sales", "support", "press").
type LeadEmail struct {
	Type  string `json:"type"`
	Email string `json:"email"`
}

// LeadPerson is a founder or key person associated with the company.
// Linkedin and X are nil when no profile URL was discovered.
type LeadPerson struct {
	Name     string  `json:"name"`
	Role     string  `json:"role"`
	LinkedIn *string `json:"linkedin,omitempty"`
	X        *string `json:"x,omitempty"`
}

// LeadData is the enriched company profile.
type LeadData struct {
	CompanyName string        `json:"company_name,omitempty"`
	Summary     string        `json:"summary,omitempty"`
	Socials     LeadSocials   `json:"socials,omitempty"`
	Tech        []string      `json:"tech,omitempty"`
	Pricing     []LeadPricing `json:"pricing,omitempty"`
	Emails      []LeadEmail   `json:"emails,omitempty"`
	People      []LeadPerson  `json:"people,omitempty"`
}

// LeadResponse contains the enriched lead for a company URL.
type LeadResponse struct {
	URL          string   `json:"url"`
	Domain       string   `json:"domain"`
	Lead         LeadData `json:"lead"`
	PeopleSource string   `json:"people_source,omitempty"`
	Cache        string   `json:"cache,omitempty"`
	Credits      int      `json:"credits,omitempty"`
}

// Lead enriches a company URL into a structured sales lead: company name,
// summary, socials, tech stack, pricing plans, categorized contact emails, and
// the founders behind the company — each with their LinkedIn and X profiles when
// discoverable. PeopleSource reports how the founder list was resolved.
//
// Billing: a flat 100 credits per successful lead.
func (c *Client) Lead(ctx context.Context, req *LeadRequest) (*LeadResponse, error) {
	var resp LeadResponse
	if err := c.do(ctx, "POST", "/v1/lead", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
