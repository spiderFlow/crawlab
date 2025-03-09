package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LLMModel represents a specific model within an LLM provider
type LLMModel struct {
	BaseModel[LLMModel] `bson:",inline"`
	ProviderId          primitive.ObjectID     `json:"provider_id" bson:"provider_id"`                         // Reference to the provider
	ModelId             string                 `json:"model_id" bson:"model_id"`                               // Provider's model ID
	Name                string                 `json:"name" bson:"name"`                                       // Model name
	DisplayName         string                 `json:"display_name" bson:"display_name"`                       // Display name for UI
	Description         string                 `json:"description,omitempty" bson:"description,omitempty"`     // Description of the model
	IsEnabled           bool                   `json:"is_enabled" bson:"is_enabled"`                           // Whether this model is enabled
	Priority            int                    `json:"priority" bson:"priority"`                               // Priority for sorting in UI
	ModelFamily         string                 `json:"model_family,omitempty" bson:"model_family,omitempty"`   // Family this model belongs to (e.g., "gpt-4", "claude")
	ContextSize         int                    `json:"context_size" bson:"context_size"`                       // Context window size in tokens
	MaxOutputTokens     int                    `json:"max_output_tokens" bson:"max_output_tokens"`             // Maximum output tokens
	SupportedFeatures   []string               `json:"supported_features" bson:"supported_features"`           // Features supported by this model
	DefaultParameters   map[string]interface{} `json:"default_parameters" bson:"default_parameters"`           // Default parameters for this model
	TokenPricing        *TokenPricing          `json:"token_pricing,omitempty" bson:"token_pricing,omitempty"` // Pricing information
}

// TokenPricing represents the pricing structure for tokens used by an LLM model
type TokenPricing struct {
	InputTokenPrice  float64 `json:"input_token_price" bson:"input_token_price"`   // Price per input token (USD per 1M tokens)
	OutputTokenPrice float64 `json:"output_token_price" bson:"output_token_price"` // Price per output token (USD per 1M tokens)
	Currency         string  `json:"currency" bson:"currency"`                     // Currency for pricing, default is USD
}

// GetModelColName returns the collection name for the model
func (m *LLMModel) GetModelColName() string {
	return "llm_models"
}

// Validate validates the model
func (m *LLMModel) Validate() error {
	// Basic validation can be implemented here
	return nil
}
