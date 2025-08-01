package reqtype

type SpaceCreateOption struct {
	Name                 string `json:"name"`
	TmplId               uint32 `json:"tmpl_id"`
	SpaceSpecId          uint32 `json:"space_spec_id"`
	UserId               uint32 `json:"user_id"`
	GitRepository        string `json:"git_repository"`
	// Anthropic API 配置
	AnthropicAuthToken   string `json:"anthropic_auth_token,omitempty"`
	AnthropicBaseURL     string `json:"anthropic_base_url,omitempty"`
	// OpenAI API 配置
	OpenAIAPIKey         string `json:"openai_api_key,omitempty"`
	OpenAIBaseURL        string `json:"openai_base_url,omitempty"`
	// DeepSeek API 配置
	DeepSeekAPIKey       string `json:"deepseek_api_key,omitempty"`
	// Gemini API 配置
	GeminiAPIKey         string `json:"gemini_api_key,omitempty"`
	// Moonshot API 配置
	MoonshotAPIKey       string `json:"moonshot_api_key,omitempty"`
	// Qwen API 配置
	QwenAPIKey           string `json:"qwen_api_key,omitempty"`
	// 模型配置
	BigModel             string `json:"big_model,omitempty"`
	SmallModel           string `json:"small_model,omitempty"`
}

type SpaceId struct {
	Id uint32 `json:"id"`
}
