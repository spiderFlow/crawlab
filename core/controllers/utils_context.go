package controllers

import (
	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/gin-gonic/gin"
)

func GetUserFromContext(c *gin.Context) (u *models.User) {
	value, ok := c.Get(constants.UserContextKey)
	if !ok {
		return nil
	}
	u, ok = value.(*models.User)
	if !ok {
		return nil
	}
	return u
}
