package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type DependencyConfigSetup struct {
	any                              `collection:"dependency_config_setups"`
	BaseModel[DependencyConfigSetup] `bson:",inline"`
	DependencyConfigId               primitive.ObjectID `json:"dependency_config_id" bson:"dependency_config_id"`
	NodeId                           primitive.ObjectID `json:"node_id" bson:"node_id"`
	Version                          string             `json:"version" bson:"version"`
	Status                           string             `json:"status" bson:"status"`
	Error                            string             `json:"error,omitempty" bson:"error,omitempty"`
}
