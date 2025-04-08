package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatMessage struct {
	any            `collection:"chat_messages"`
	BaseModel      `bson:",inline"`
	ConversationId primitive.ObjectID   `json:"conversation_id" bson:"conversation_id" description:"Conversation ID"`
	Role           string               `json:"role" bson:"role" description:"Message role (user/assistant/system)"`
	Content        string               `json:"content,omitempty" bson:"content,omitempty" description:"Message content"`
	IsAgent        bool                 `json:"is_agent,omitempty" bson:"is_agent,omitempty" description:"Is agent"`
	ContentIds     []primitive.ObjectID `json:"content_ids,omitempty" bson:"content_ids,omitempty" description:"Content IDs"`
	Contents       []ChatMessageContent `json:"contents,omitempty" bson:"-" description:"Contents"`
	Tokens         int                  `json:"tokens" bson:"tokens" description:"Number of tokens in the message"`
	Model          string               `json:"model" bson:"model" description:"AI model used"`
	Metadata       map[string]any       `json:"metadata,omitempty" bson:"metadata,omitempty" description:"Additional metadata"`
	Status         string               `json:"status" bson:"status" description:"Message status (pending/completed/failed)"`
	Error          string               `json:"error,omitempty" bson:"error,omitempty" description:"Error message if failed"`
}

func (m *ChatMessage) GetContent() string {
	// If the message has a single content item, return it directly
	if m.Content != "" {
		return m.Content
	}

	// If the message has multiple content items, concatenate them
	var result string
	for _, content := range m.Contents {
		switch content.Type {
		case ChatMessageContentTypeText:
			result += content.Content
		case ChatMessageContentTypeAction:
			// Format action content with status
			actionInfo := "[" + content.Action
			if content.ActionStatus != "" {
				actionInfo += " - " + string(content.ActionStatus)
			}
			actionInfo += "]"

			if content.Content != "" {
				result += actionInfo + ": " + content.Content
			} else {
				result += actionInfo
			}
		default:
			// For any unrecognized type, just add the content
			result += content.Content
		}

		// Add newline between content items
		result += "\n\n"
	}

	return result
}
