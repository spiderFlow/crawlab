package controllers

import (
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/user"
	"github.com/gin-gonic/gin"
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
