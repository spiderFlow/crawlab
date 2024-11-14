package models

type Role struct {
	any             `collection:"roles"`
	BaseModel[Role] `bson:",inline"`
	Name            string   `json:"name" bson:"name"`
	Description     string   `json:"description" bson:"description"`
	Routes          []string `json:"routes" bson:"routes"`
	RootAdmin       bool     `json:"-" bson:"root_admin,omitempty"`
	IsRootAdmin     bool     `json:"root_admin" bson:"-"`
	Users           int      `json:"users" bson:"-"`
}
