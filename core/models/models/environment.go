package models

type Environment struct {
	any       `collection:"environments"`
	BaseModel `bson:",inline"`
	Key       string `json:"key" bson:"key"`
	Value     string `json:"value" bson:"value"`
}
