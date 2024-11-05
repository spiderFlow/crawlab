package controllers

import (
	"errors"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetSetting(c *gin.Context) {
	// key
	key := c.Param("id")

	// setting
	s, err := service.NewModelService[models.Setting]().GetOne(bson.M{"key": key}, nil)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			HandleSuccess(c)
			return
		}
		HandleErrorInternalServerError(c, err)
		return
	}

	HandleSuccessWithData(c, s)
}

func PostSetting(c *gin.Context) {
	// key
	key := c.Param("id")

	// settings
	var s models.Setting
	if err := c.ShouldBindJSON(&s); err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	if s.Key == "" {
		s.Key = key
	}

	u := GetUserFromContext(c)

	s.SetCreated(u.Id)
	s.SetUpdated(u.Id)

	id, err := service.NewModelService[models.Setting]().InsertOne(s)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	s.Id = id

	HandleSuccessWithData(c, s)
}

func PutSetting(c *gin.Context) {
	// key
	key := c.Param("id")

	// settings
	var s models.Setting
	if err := c.ShouldBindJSON(&s); err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	modelSvc := service.NewModelService[models.Setting]()

	// setting
	_s, err := modelSvc.GetOne(bson.M{"key": key}, nil)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	u := GetUserFromContext(c)

	// save
	_s.Value = s.Value
	_s.SetUpdated(u.Id)
	err = modelSvc.ReplaceOne(bson.M{"key": key}, *_s)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	HandleSuccess(c)
}
