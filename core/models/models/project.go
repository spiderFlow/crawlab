package models

type Project struct {
	any         `collection:"projects"`
	BaseModel   `bson:",inline"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	Spiders     int    `json:"spiders" bson:"-"`
}
