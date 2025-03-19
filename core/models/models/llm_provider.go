package models

// LLMProvider represents a language model provider such as OpenAI, Anthropic, etc.
type LLMProvider struct {
	any            `collection:"llm_providers"`
	BaseModel      `bson:",inline"`
	Key            string   `json:"key" bson:"key" description:"Provider key (e.g., 'openai', 'anthropic', 'gemini')"`
	Name           string   `json:"name" bson:"name" description:"Display name for UI"`
	Enabled        bool     `json:"enabled" bson:"enabled" description:"Whether this provider is enabled"`
	ApiKey         string   `json:"api_key" bson:"api_key" description:"API key for the provider"`
	ApiBaseUrl     string   `json:"api_base_url" bson:"api_base_url" description:"API base URL for the provider"`
	DeploymentName string   `json:"deployment_name" bson:"deployment_name" description:"Deployment name for the provider"`
	ApiVersion     string   `json:"api_version" bson:"api_version" description:"API version for the provider"`
	Models         []string `json:"models" bson:"models" description:"Models supported by this provider"`
	Unset          bool     `json:"unset" bson:"-"` // Whether the provider is unset
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
