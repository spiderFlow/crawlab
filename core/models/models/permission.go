package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Permission struct {
	any         `collection:"permissions"`
	BaseModel   `bson:",inline"`
	Key         string             `json:"key" bson:"key"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	RoleId      primitive.ObjectID `json:"role_id" bson:"role_id"`
	Type        string             `json:"type" bson:"type"`
	Routes      []string           `json:"routes" bson:"routes"`
}
