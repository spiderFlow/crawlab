package models

type Role struct {
	any             `collection:"roles"`
	BaseModel[Role] `bson:",inline"`
	Key             string `json:"key" bson:"key"`
	Name            string `json:"name" bson:"name"`
	Description     string `json:"description" bson:"description"`
}
