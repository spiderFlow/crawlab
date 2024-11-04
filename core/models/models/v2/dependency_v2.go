package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DependencyV2 struct {
	any                       `collection:"dependencies"`
	BaseModelV2[DependencyV2] `bson:",inline"`
	Name                      string               `json:"name" bson:"name"`
	Description               string               `json:"description" bson:"description"`
	NodeId                    primitive.ObjectID   `json:"node_id" bson:"node_id"`
	Type                      string               `json:"type" bson:"type"`
	Version                   string               `json:"version" bson:"version"`
	Status                    string               `json:"status" bson:"status"`
	Error                     string               `json:"error,omitempty" bson:"error,omitempty"`
	Logs                      []string             `json:"logs,omitempty" bson:"logs,omitempty"`
	NodeIds                   []primitive.ObjectID `json:"node_ids,omitempty" bson:"-"`
	Versions                  []string             `json:"versions,omitempty" bson:"-"`
}
