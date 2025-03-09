package models

// LLMProvider represents a language model provider such as OpenAI, Anthropic, etc.
type LLMProvider struct {
	any                    `collection:"llm_providers"`
	BaseModel[LLMProvider] `bson:",inline"`
	Key                    string   `json:"key" bson:"key"`                                     // Provider key (e.g., "openai", "anthropic", "gemini")
	Name                   string   `json:"name" bson:"name"`                                   // Display name for UI
	Description            string   `json:"description,omitempty" bson:"description,omitempty"` // Description of the provider
	Models                 []string `json:"models" bson:"models"`                               // Models supported by this provider
	ApiKey                 string   `json:"api_key" bson:"api_key"`                             // API key for the provider
	ApiBaseUrl             string   `json:"api_base_url" bson:"api_base_url"`                   // API base URL for the provider
	Enabled                bool     `json:"enabled" bson:"enabled"`                             // Whether this provider is enabled
	Priority               int      `json:"priority" bson:"priority"`                           // Priority for sorting in UI
	ConfigSchema           string   `json:"config_schema" bson:"config_schema"`                 // JSON schema for configuration
	DefaultConfig          string   `json:"default_config" bson:"default_config"`               // Default configuration as JSON
	SupportedFeatures      []string `json:"supported_features" bson:"supported_features"`       // Features supported by this provider (e.g., "function_calling", "streaming")
}
