package controllers

import (
	"errors"
	"fmt"
	"github.com/crawlab-team/crawlab/core/mongo"
	"regexp"

	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
)

func GetUserById(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}
	getUserById(id, c)
}

func GetUserList(c *gin.Context) {
	// params
	pagination := MustGetPagination(c)
	query := MustGetFilterQuery(c)
	sort := MustGetSortOption(c)

	// get users
	users, err := service.NewModelService[models.User]().GetMany(query, &mongo.FindOptions{
		Sort:  sort,
		Skip:  pagination.Size * (pagination.Page - 1),
		Limit: pagination.Size,
	})
	if err != nil {
		if errors.Is(err, mongo2.ErrNoDocuments) {
			HandleSuccessWithListData(c, nil, 0)
		} else {
			HandleErrorInternalServerError(c, err)
		}
		return
	}

	// get roles
	if utils.IsPro() {
		var roleIds []primitive.ObjectID
		for _, user := range users {
			if !user.RoleId.IsZero() {
				roleIds = append(roleIds, user.RoleId)
			}
		}
		if len(roleIds) > 0 {
			roles, err := service.NewModelService[models.Role]().GetMany(bson.M{
				"_id": bson.M{"$in": roleIds},
			}, nil)
			if err != nil {
				HandleErrorInternalServerError(c, err)
				return
			}
			rolesMap := make(map[primitive.ObjectID]models.Role)
			for _, role := range roles {
				rolesMap[role.Id] = role
			}
			for i, user := range users {
				if user.RoleId.IsZero() {
					continue
				}
				if role, ok := rolesMap[user.RoleId]; ok {
					users[i].Role = role.Name
					users[i].RootAdminRole = role.RootAdmin
				}
			}
		}
	}

	// total count
	total, err := service.NewModelService[models.User]().Count(query)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// response
	HandleSuccessWithListData(c, users, total)
}

func PostUser(c *gin.Context) {
	var payload struct {
		Username string             `json:"username"`
		Password string             `json:"password"`
		Role     string             `json:"role"`
		RoleId   primitive.ObjectID `json:"role_id"`
		Email    string             `json:"email"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// Validate email format
	if payload.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(payload.Email) {
			HandleErrorBadRequest(c, fmt.Errorf("invalid email format"))
			return
		}
	}

	if !payload.RoleId.IsZero() {
		_, err := service.NewModelService[models.Role]().GetById(payload.RoleId)
		if err != nil {
			HandleErrorBadRequest(c, err)
			return
		}
	}
	u := GetUserFromContext(c)
	model := models.User{
		Username: payload.Username,
		Password: utils.EncryptMd5(payload.Password),
		Role:     payload.Role,
		RoleId:   payload.RoleId,
		Email:    payload.Email,
	}
	model.SetCreated(u.Id)
	model.SetUpdated(u.Id)
	id, err := service.NewModelService[models.User]().InsertOne(model)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	result, err := service.NewModelService[models.User]().GetById(id)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	HandleSuccessWithData(c, result)
}

func PutUserById(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}
	putUser(id, c)
}

func PostUserChangePassword(c *gin.Context) {
	// get id
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	postUserChangePassword(id, c)
}

func DeleteUserById(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	user, err := service.NewModelService[models.User]().GetById(id)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	if user.RootAdmin {
		HandleErrorForbidden(c, errors.New("root admin cannot be deleted"))
		return
	}

	if err := service.NewModelService[models.User]().DeleteById(id); err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	HandleSuccess(c)
}

func DeleteUserList(c *gin.Context) {
	type Payload struct {
		Ids []string `json:"ids"`
	}

	var payload Payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// Convert string IDs to ObjectIDs
	var ids []primitive.ObjectID
	for _, id := range payload.Ids {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			HandleErrorBadRequest(c, err)
			return
		}
		ids = append(ids, objectId)
	}

	// Check if root admin is in the list
	_, err := service.NewModelService[models.User]().GetOne(bson.M{
		"_id": bson.M{
			"$in": ids,
		},
		"root_admin": true,
	}, nil)
	if err == nil {
		HandleErrorForbidden(c, errors.New("root admin cannot be deleted"))
		return
	}
	if !errors.Is(err, mongo2.ErrNoDocuments) {
		HandleErrorInternalServerError(c, err)
		return
	}

	// Delete users
	if err := service.NewModelService[models.User]().DeleteMany(bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}); err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	HandleSuccess(c)
}

func GetUserMe(c *gin.Context) {
	u := GetUserFromContext(c)
	getUserByIdWithRoutes(u.Id, c)
}

func PutUserMe(c *gin.Context) {
	u := GetUserFromContext(c)
	putUser(u.Id, c)
}

func PostUserMeChangePassword(c *gin.Context) {
	u := GetUserFromContext(c)
	postUserChangePassword(u.Id, c)
}

func getUserById(userId primitive.ObjectID, c *gin.Context) {
	// get user
	user, err := service.NewModelService[models.User]().GetById(userId)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// get role
	if utils.IsPro() {
		if !user.RoleId.IsZero() {
			role, err := service.NewModelService[models.Role]().GetById(user.RoleId)
			if err != nil {
				HandleErrorInternalServerError(c, err)
				return
			}
			user.Role = role.Name
			user.RootAdminRole = role.RootAdmin
		}
	}

	HandleSuccessWithData(c, user)
}

func getUserByIdWithRoutes(userId primitive.ObjectID, c *gin.Context) {
	if !utils.IsPro() {
		getUserById(userId, c)
		return
	}

	// get user
	user, err := service.NewModelService[models.User]().GetById(userId)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// get role
	if !user.RoleId.IsZero() {
		role, err := service.NewModelService[models.Role]().GetById(user.RoleId)
		if err != nil {
			HandleErrorInternalServerError(c, err)
			return
		}
		user.Role = role.Name
		user.RootAdminRole = role.RootAdmin
		user.Routes = role.Routes
	}

	HandleSuccessWithData(c, user)
}

func putUser(userId primitive.ObjectID, c *gin.Context) {
	// get payload
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// model service
	modelSvc := service.NewModelService[models.User]()

	// update user
	userDb, err := modelSvc.GetById(userId)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// if root admin, disallow changing username and role
	if userDb.RootAdmin {
		user.Username = userDb.Username
		user.RoleId = userDb.RoleId
	}

	// disallow changing password
	user.Password = userDb.Password

	// current user
	u := GetUserFromContext(c)

	// update user
	user.SetUpdated(u.Id)
	if user.Id.IsZero() {
		user.Id = userId
	}
	if err := modelSvc.ReplaceById(userId, user); err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// handle success
	HandleSuccess(c)
}

func postUserChangePassword(userId primitive.ObjectID, c *gin.Context) {
	// get payload
	var payload struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}
	if len(payload.Password) < 5 {
		HandleErrorBadRequest(c, errors.New("password must be at least 5 characters"))
		return
	}

	// current user
	u := GetUserFromContext(c)

	// update password
	userDb, err := service.NewModelService[models.User]().GetById(userId)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	userDb.SetUpdated(u.Id)
	userDb.Password = utils.EncryptMd5(payload.Password)
	if err := service.NewModelService[models.User]().ReplaceById(userDb.Id, *userDb); err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// handle success
	HandleSuccess(c)
}
