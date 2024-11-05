package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type DependencyLog struct {
	any                      `collection:"dependency_logs"`
	BaseModel[DependencyLog] `bson:",inline"`
	DependencyId             primitive.ObjectID `json:"dependency_id" bson:"dependency_id"`
	Content                  string             `json:"content" bson:"content"`
}
