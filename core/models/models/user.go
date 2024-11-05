package models

type User struct {
	any             `collection:"users"`
	BaseModel[User] `bson:",inline"`
	Username        string `json:"username" bson:"username"`
	Password        string `json:"-,omitempty" bson:"password"`
	Role            string `json:"role" bson:"role"`
	Email           string `json:"email" bson:"email"`
}
