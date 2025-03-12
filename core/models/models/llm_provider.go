package models

// LLMProvider represents a language model provider such as OpenAI, Anthropic, etc.
type LLMProvider struct {
	any            `collection:"llm_providers"`
	BaseModel      `bson:",inline"`
	Key            string   `json:"key" bson:"key"`                         // Provider key (e.g., "openai", "anthropic", "gemini")
	Name           string   `json:"name" bson:"name"`                       // Display name for UI
	Enabled        bool     `json:"enabled" bson:"enabled"`                 // Whether this provider is enabled
	ApiKey         string   `json:"api_key" bson:"api_key"`                 // API key for the provider
	ApiBaseUrl     string   `json:"api_base_url" bson:"api_base_url"`       // API base URL for the provider
	DeploymentName string   `json:"deployment_name" bson:"deployment_name"` // Deployment name for the provider
	ApiVersion     string   `json:"api_version" bson:"api_version"`         // API version for the provider
	Models         []string `json:"models" bson:"models"`                   // Models supported by this provider
	Unset          bool     `json:"unset" bson:"-"`                         // Whether the provider is unset
}

func (p *LLMProvider) IsUnset() bool {
	if p.ApiKey == "" {
		return true
	}
	if len(p.Models) == 0 {
		return true
	}
	switch p.Key {
	case "azure-openai":
		return p.ApiBaseUrl == "" || p.DeploymentName == "" || p.ApiVersion == ""
	case "openai-compatible":
		return p.ApiBaseUrl == ""
	}
	return false
}
