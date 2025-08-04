package reqtype

type SpaceCreateOption struct {
	Name                 string `json:"name"`
	TmplId               uint32 `json:"tmpl_id"`
	SpaceSpecId          uint32 `json:"space_spec_id"`
	UserId               uint32 `json:"user_id"`
	GitRepository        string `json:"git_repository"`
	// Claude API 配置 - 简化版
	AnthropicAuthToken   string `json:"anthropic_auth_token,omitempty"`
	AnthropicBaseURL     string `json:"anthropic_base_url,omitempty"`
}

type SpaceId struct {
	Id uint32 `json:"id"`
}
