package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type DependencyLog struct {
	any       `collection:"dependency_logs"`
	BaseModel `bson:",inline"`
	TargetId  primitive.ObjectID `json:"target_id" bson:"target_id"`
	Content   string             `json:"content" bson:"content"`
}
