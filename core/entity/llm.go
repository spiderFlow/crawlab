package entity

type LLMResponseUsage struct {
	InputTokens  int `json:"input_tokens" bson:"input_tokens"`
	OutputTokens int `json:"output_tokens" bson:"output_tokens"`
	TotalTokens  int `json:"total_tokens" bson:"total_tokens"`
}
