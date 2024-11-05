package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type DependencyLog struct {
	any                      `collection:"dependency_logs"`
	BaseModel[DependencyLog] `bson:",inline"`
	TaskId                   primitive.ObjectID `json:"task_id" bson:"task_id"`
	Content                  string             `json:"content" bson:"content"`
}
