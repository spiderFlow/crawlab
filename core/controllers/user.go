package controllers

import (
	"errors"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/db/mongo"
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
					users[i].IsAdmin = role.Admin
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

func GetUserMe(c *gin.Context) {
	u := GetUserFromContext(c)
	getUserById(u.Id, c)
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
			user.IsAdmin = role.Admin
		}
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
		user.Id = u.Id
	}
	if err := modelSvc.ReplaceById(u.Id, user); err != nil {
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
