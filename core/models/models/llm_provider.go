package models

// LLMProvider represents a language model provider such as OpenAI, Anthropic, etc.
type LLMProvider struct {
	BaseModel[LLMProvider] `bson:",inline"`
	Name                   string   `json:"name" bson:"name"`                                   // Provider name (e.g., "openai", "anthropic", "gemini")
	DisplayName            string   `json:"display_name" bson:"display_name"`                   // Display name for UI
	Description            string   `json:"description,omitempty" bson:"description,omitempty"` // Description of the provider
	IsEnabled              bool     `json:"is_enabled" bson:"is_enabled"`                       // Whether this provider is enabled
	Priority               int      `json:"priority" bson:"priority"`                           // Priority for sorting in UI
	ConfigSchema           string   `json:"config_schema" bson:"config_schema"`                 // JSON schema for configuration
	DefaultConfig          string   `json:"default_config" bson:"default_config"`               // Default configuration as JSON
	SupportedFeatures      []string `json:"supported_features" bson:"supported_features"`       // Features supported by this provider (e.g., "function_calling", "streaming")
}

// GetModelColName returns the collection name for the provider model
func (p *LLMProvider) GetModelColName() string {
	return "llm_providers"
}

// Validate validates the provider model
func (p *LLMProvider) Validate() error {
	// Basic validation can be implemented here
	return nil
}
