package models

type Role struct {
	any             `collection:"roles"`
	BaseModel[Role] `bson:",inline"`
	Name            string   `json:"name" bson:"name"`
	Description     string   `json:"description" bson:"description"`
	Routes          []string `json:"routes" bson:"routes"`
	Admin           bool     `json:"-" bson:"admin,omitempty"`
	IsAdmin         bool     `json:"admin" bson:"-"`
	Users           int      `json:"users" bson:"-"`
}
