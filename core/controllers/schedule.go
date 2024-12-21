package controllers

import (
	errors2 "errors"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/schedule"
	"github.com/crawlab-team/crawlab/core/spider/admin"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PostSchedule(c *gin.Context) {
	var s models.Schedule
	if err := c.ShouldBindJSON(&s); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	u := GetUserFromContext(c)

	modelSvc := service.NewModelService[models.Schedule]()

	s.SetCreated(u.Id)
	s.SetUpdated(u.Id)
	id, err := modelSvc.InsertOne(s)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	s.Id = id

	if s.Enabled {
		scheduleSvc := schedule.GetScheduleService()
		if err := scheduleSvc.Enable(s, u.Id); err != nil {
			HandleErrorInternalServerError(c, err)
			return
		}
	}

	HandleSuccessWithData(c, s)
}

func PutScheduleById(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}
	var s models.Schedule
	if err := c.ShouldBindJSON(&s); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}
	if s.Id != id {
		HandleErrorBadRequest(c, errors2.New("id in path does not match id in body"))
		return
	}

	modelSvc := service.NewModelService[models.Schedule]()
	err = modelSvc.ReplaceById(id, s)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	scheduleSvc := schedule.GetScheduleService()

	u := GetUserFromContext(c)

	if s.Enabled {
		if err := scheduleSvc.Enable(s, u.Id); err != nil {
			HandleErrorInternalServerError(c, err)
			return
		}
	} else {
		if err := scheduleSvc.Disable(s, u.Id); err != nil {
			HandleErrorInternalServerError(c, err)
			return
		}
	}

	HandleSuccessWithData(c, s)
}

func PostScheduleEnable(c *gin.Context) {
	postScheduleEnableDisableFunc(true)(c)
}

func PostScheduleDisable(c *gin.Context) {
	postScheduleEnableDisableFunc(false)(c)
}

func postScheduleEnableDisableFunc(isEnable bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			HandleErrorBadRequest(c, err)
			return
		}
		svc := schedule.GetScheduleService()
		s, err := service.NewModelService[models.Schedule]().GetById(id)
		if err != nil {
			HandleErrorInternalServerError(c, err)
			return
		}
		u := GetUserFromContext(c)
		if isEnable {
			err = svc.Enable(*s, u.Id)
		} else {
			err = svc.Disable(*s, u.Id)
		}
		if err != nil {
			HandleErrorInternalServerError(c, err)
			return
		}
		HandleSuccess(c)
	}
}

func PostScheduleRun(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// options
	var opts interfaces.SpiderRunOptions
	if err := c.ShouldBindJSON(&opts); err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	if opts.ScheduleId.IsZero() {
		opts.ScheduleId = id
	}

	// schedule
	sch, err := service.NewModelService[models.Schedule]().GetById(id)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// spider
	s, err := service.NewModelService[models.Spider]().GetById(sch.SpiderId)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// user
	if u := GetUserFromContext(c); u != nil {
		opts.UserId = u.GetId()
	}

	// schedule tasks
	taskIds, err := admin.GetSpiderAdminService().Schedule(s.Id, &opts)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	HandleSuccessWithData(c, taskIds)
}
