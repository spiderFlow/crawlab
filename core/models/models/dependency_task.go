package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type DependencyTask struct {
	any                       `collection:"dependency_tasks"`
	BaseModel[DependencyTask] `bson:",inline"`
	Status                    string             `json:"status" bson:"status"`
	Error                     string             `json:"error" bson:"error"`
	SettingId                 primitive.ObjectID `json:"setting_id" bson:"setting_id"`
	Type                      string             `json:"type" bson:"type"`
	NodeId                    primitive.ObjectID `json:"node_id" bson:"node_id"`
	Action                    string             `json:"action" bson:"action"`
	DepNames                  []string           `json:"dep_names" bson:"dep_names"`
}
