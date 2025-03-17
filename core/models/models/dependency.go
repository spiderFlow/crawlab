package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Dependency struct {
	any         `collection:"dependencies"`
	BaseModel   `bson:",inline"`
	Name        string               `json:"name" bson:"name"`
	Description string               `json:"description" bson:"description"`
	NodeId      primitive.ObjectID   `json:"node_id" bson:"node_id"`
	Type        string               `json:"type" bson:"type"`
	Version     string               `json:"version" bson:"version"`
	Status      string               `json:"status" bson:"status"`
	Error       string               `json:"error,omitempty" bson:"error,omitempty"`
	NodeIds     []primitive.ObjectID `json:"node_ids,omitempty" bson:"-"`
	Versions    []string             `json:"versions,omitempty" bson:"-"`
}
