package webclaw

import "context"

// AgentScrapeRequest configures a goal-directed agent scrape.
type AgentScrapeRequest struct {
	URL      string `json:"url"`
	Goal     string `json:"goal"`
	MaxSteps int    `json:"max_steps,omitempty"`
}

// AgentStep describes one action taken by the agent.
type AgentStep struct {
	Step   int         `json:"step"`
	Action interface{} `json:"action"`
}

// AgentScrapeResponse is the result of an agent scrape.
type AgentScrapeResponse struct {
	Data       map[string]interface{} `json:"data"`
	Steps      []AgentStep            `json:"steps"`
	URL        string                 `json:"url"`
	TotalSteps int                    `json:"total_steps"`
	Warning    string                 `json:"warning,omitempty"`
}

// AgentScrape performs a goal-directed scrape using an AI agent.
func (c *Client) AgentScrape(ctx context.Context, req *AgentScrapeRequest) (*AgentScrapeResponse, error) {
	var resp AgentScrapeResponse
	if err := c.do(ctx, "POST", "/v1/agent-scrape", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
