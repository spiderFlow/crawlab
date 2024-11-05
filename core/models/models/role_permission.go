package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RolePermission struct {
	any                       `collection:"role_permissions"`
	BaseModel[RolePermission] `bson:",inline"`
	RoleId                    primitive.ObjectID `json:"role_id" bson:"role_id"`
	PermissionId              primitive.ObjectID `json:"permission_id" bson:"permission_id"`
}
