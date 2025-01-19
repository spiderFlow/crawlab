package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	any             `collection:"users"`
	BaseModel[User] `bson:",inline"`
	Username        string             `json:"username" bson:"username"`
	Password        string             `json:"-" bson:"password"`
	Role            string             `json:"role" bson:"role"`
	RoleId          primitive.ObjectID `json:"role_id" bson:"role_id"`
	FirstName       string             `json:"first_name" bson:"first_name"`
	LastName        string             `json:"last_name" bson:"last_name"`
	Email           string             `json:"email" bson:"email"`
	RootAdmin       bool               `json:"root_admin,omitempty" bson:"root_admin"`
	RootAdminRole   bool               `json:"root_admin_role,omitempty" bson:"-"`
	Routes          []string           `json:"routes,omitempty" bson:"-"`
}
