package models

type Project struct {
	any         `collection:"projects"`
	BaseModel   `bson:",inline"`
	Name        string `json:"name" bson:"name" description:"Name"`
	Description string `json:"description" bson:"description" description:"Description"`
	Spiders     int    `json:"spiders" bson:"-"`
}
