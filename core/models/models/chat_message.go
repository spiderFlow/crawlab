package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatMessage struct {
	any            `collection:"chat_messages"`
	BaseModel      `bson:",inline"`
	ConversationId primitive.ObjectID `json:"conversation_id" bson:"conversation_id" description:"Conversation ID"`
	Role           string             `json:"role" bson:"role" description:"Message role (user/assistant/system)"`
	Content        string             `json:"content" bson:"content" description:"Message content"`
	Tokens         int                `json:"tokens" bson:"tokens" description:"Number of tokens in the message"`
	Model          string             `json:"model" bson:"model" description:"AI model used"`
	Metadata       map[string]any     `json:"metadata,omitempty" bson:"metadata,omitempty" description:"Additional metadata"`
	Status         string             `json:"status" bson:"status" description:"Message status (pending/completed/failed)"`
	Error          string             `json:"error,omitempty" bson:"error,omitempty" description:"Error message if failed"`
}
