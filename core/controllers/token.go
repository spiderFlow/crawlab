package controllers

import (
	"errors"

	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/mongo"
	"github.com/crawlab-team/crawlab/core/user"
	"github.com/gin-gonic/gin"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
)

func PostToken(c *gin.Context) {
	var t models.Token
	if err := c.ShouldBindJSON(&t); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}
	svc, err := user.GetUserService()
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	u := GetUserFromContext(c)
	t.SetCreated(u.Id)
	t.SetUpdated(u.Id)
	t.Token, err = svc.MakeToken(u)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	_, err = service.NewModelService[models.Token]().InsertOne(t)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	HandleSuccess(c)
}

func GetTokenList(c *gin.Context) {
	// Get current user from context
	u := GetUserFromContext(c)

	// Get pagination, filter query, and sort options
	pagination := MustGetPagination(c)
	query := MustGetFilterQuery(c)
	sort := MustGetSortOption(c)

	// If query is nil, initialize it
	if query == nil {
		query = make(map[string]interface{})
	}

	// Add filter for tokens created by the current user
	query["created_by"] = u.Id

	// Get tokens with pagination
	tokens, err := service.NewModelService[models.Token]().GetMany(query, &mongo.FindOptions{
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

	// Count total tokens for pagination
	total, err := service.NewModelService[models.Token]().Count(query)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// Return tokens with total count
	HandleSuccessWithListData(c, tokens, total)
}
