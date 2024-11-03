package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type DependencyRepo struct {
	Name          string               `json:"name,omitempty" bson:"name,omitempty"`
	NodeIds       []primitive.ObjectID `json:"node_ids,omitempty" bson:"node_ids,omitempty"`
	Versions      []string             `json:"versions,omitempty" bson:"versions,omitempty"`
	LatestVersion string               `json:"latest_version" bson:"latest_version"`
	Type          string               `json:"type" bson:"type"`
}
