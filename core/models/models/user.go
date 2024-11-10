package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	any             `collection:"users"`
	BaseModel[User] `bson:",inline"`
	Username        string             `json:"username" bson:"username"`
	Password        string             `json:"-" bson:"password"`
	Role            string             `json:"role" bson:"role"`
	RoleId          primitive.ObjectID `json:"role_id" bson:"role_id"`
	Email           string             `json:"email" bson:"email"`
	IsAdmin         bool               `json:"admin" bson:"-"`
}
