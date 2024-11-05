package models

import (
	"go.mongodb.org/mongo-driver/bson"
)

type Setting struct {
	any                `collection:"settings"`
	BaseModel[Setting] `bson:",inline"`
	Key                string `json:"key" bson:"key"`
	Value              bson.M `json:"value" bson:"value"`
}
