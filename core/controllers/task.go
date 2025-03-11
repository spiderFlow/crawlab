package controllers

import (
	"errors"
	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	mongo3 "github.com/crawlab-team/crawlab/core/mongo"
	"github.com/crawlab-team/crawlab/core/spider/admin"
	"github.com/crawlab-team/crawlab/core/task/log"
	"github.com/crawlab-team/crawlab/core/task/scheduler"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func GetTaskById(c *gin.Context) {
	// id
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// task
	t, err := service.NewModelService[models.Task]().GetById(id)
	if errors.Is(err, mongo2.ErrNoDocuments) {
		HandleErrorNotFound(c, err)
		return
	}
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// skip if task status is pending
	if t.Status == constants.TaskStatusPending {
		HandleSuccessWithData(c, t)
		return
	}

	// spider
	if !t.SpiderId.IsZero() {
		t.Spider, _ = service.NewModelService[models.Spider]().GetById(t.SpiderId)
	}

	// schedule
	if !t.ScheduleId.IsZero() {
		t.Schedule, _ = service.NewModelService[models.Schedule]().GetById(t.ScheduleId)
	}

	// node
	if !t.NodeId.IsZero() {
		t.Node, _ = service.NewModelService[models.Node]().GetById(t.NodeId)
	}

	// task stat
	t.Stat, _ = service.NewModelService[models.TaskStat]().GetById(id)

	HandleSuccessWithData(c, t)
}

func GetTaskList(c *gin.Context, params *GetListParams) {
	withStats := c.Query("stats")
	if withStats == "" {
		NewController[models.Task]().GetList(c, params)
		return
	}

	// params
	pagination := MustGetPagination(c)
	query := MustGetFilterQuery(c)
	sort := MustGetSortOption(c)

	// get tasks
	tasks, err := service.NewModelService[models.Task]().GetMany(query, &mongo3.FindOptions{
		Sort:  sort,
		Skip:  pagination.Size * (pagination.Page - 1),
		Limit: pagination.Size,
	})
	if err != nil {
		if errors.Is(err, mongo2.ErrNoDocuments) {
			HandleErrorNotFound(c, err)
		} else {
			HandleErrorInternalServerError(c, err)
		}
		return
	}

	// check empty list
	if len(tasks) == 0 {
		HandleSuccessWithListData(c, nil, 0)
		return
	}

	// ids
	var taskIds []primitive.ObjectID
	var spiderIds []primitive.ObjectID
	for _, t := range tasks {
		taskIds = append(taskIds, t.Id)
		spiderIds = append(spiderIds, t.SpiderId)
	}

	// total count
	total, err := service.NewModelService[models.Task]().Count(query)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// stat list
	stats, err := service.NewModelService[models.TaskStat]().GetMany(bson.M{
		"_id": bson.M{
			"$in": taskIds,
		},
	}, nil)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// cache stat list to dict
	statsDict := map[primitive.ObjectID]models.TaskStat{}
	for _, s := range stats {
		statsDict[s.Id] = s
	}

	// spider list
	spiders, err := service.NewModelService[models.Spider]().GetMany(bson.M{
		"_id": bson.M{
			"$in": spiderIds,
		},
	}, nil)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// cache spider list to dict
	spiderDict := map[primitive.ObjectID]models.Spider{}
	for _, s := range spiders {
		spiderDict[s.Id] = s
	}

	// iterate list again
	for i, t := range tasks {
		// task stat
		ts, ok := statsDict[t.Id]
		if ok {
			tasks[i].Stat = &ts
		}

		// spider
		s, ok := spiderDict[t.SpiderId]
		if ok {
			tasks[i].Spider = &s
		}
	}

	// response
	HandleSuccessWithListData(c, tasks, total)
}

func DeleteTaskById(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// delete in db
	if err := mongo3.RunTransaction(func(context mongo2.SessionContext) (err error) {
		// delete task
		_, err = service.NewModelService[models.Task]().GetById(id)
		if err != nil {
			return err
		}
		err = service.NewModelService[models.Task]().DeleteById(id)
		if err != nil {
			return err
		}

		// delete task stat
		_, err = service.NewModelService[models.TaskStat]().GetById(id)
		if err != nil {
			logger.Warnf("delete task stat error: %s", err.Error())
			return nil
		}
		err = service.NewModelService[models.TaskStat]().DeleteById(id)
		if err != nil {
			logger.Warnf("delete task stat error: %s", err.Error())
			return nil
		}

		return nil
	}); err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// delete task logs
	logPath := filepath.Join(utils.GetTaskLogPath(), id.Hex())
	if err := os.RemoveAll(logPath); err != nil {
		logger.Warnf("failed to remove task log directory: %s", logPath)
	}

	HandleSuccess(c)
}

func DeleteList(c *gin.Context) {
	var payload struct {
		Ids []primitive.ObjectID `json:"ids"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	if err := mongo3.RunTransaction(func(context mongo2.SessionContext) error {
		// delete tasks
		if err := service.NewModelService[models.Task]().DeleteMany(bson.M{
			"_id": bson.M{
				"$in": payload.Ids,
			},
		}); err != nil {
			return err
		}

		// delete task stats
		if err := service.NewModelService[models.Task]().DeleteMany(bson.M{
			"_id": bson.M{
				"$in": payload.Ids,
			},
		}); err != nil {
			logger.Warnf("delete task stat error: %s", err.Error())
			return nil
		}

		return nil
	}); err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// delete tasks logs
	wg := sync.WaitGroup{}
	wg.Add(len(payload.Ids))
	for _, id := range payload.Ids {
		go func(id string) {
			// delete task logs
			logPath := filepath.Join(utils.GetTaskLogPath(), id)
			if err := os.RemoveAll(logPath); err != nil {
				logger.Warnf("failed to remove task log directory: %s", logPath)
			}
			wg.Done()
		}(id.Hex())
	}
	wg.Wait()

	HandleSuccess(c)
}

func PostTaskRun(c *gin.Context) {
	// task
	var t models.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// validate spider id
	if t.SpiderId.IsZero() {
		HandleErrorBadRequest(c, errors.New("spider id is required"))
		return
	}

	// spider
	s, err := service.NewModelService[models.Spider]().GetById(t.SpiderId)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// options
	opts := &interfaces.SpiderRunOptions{
		Mode:     t.Mode,
		NodeIds:  t.NodeIds,
		Cmd:      t.Cmd,
		Param:    t.Param,
		Priority: t.Priority,
	}

	// user
	if u := GetUserFromContext(c); u != nil {
		opts.UserId = u.Id
	}

	// run
	adminSvc := admin.GetSpiderAdminService()
	taskIds, err := adminSvc.Schedule(s.Id, opts)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	HandleSuccessWithData(c, taskIds)

}

func PostTaskRestart(c *gin.Context) {
	// id
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// task
	t, err := service.NewModelService[models.Task]().GetById(id)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// options
	opts := &interfaces.SpiderRunOptions{
		Mode:     t.Mode,
		NodeIds:  t.NodeIds,
		Cmd:      t.Cmd,
		Param:    t.Param,
		Priority: t.Priority,
	}

	// normalize options for tasks with assigned node
	if !t.NodeId.IsZero() {
		opts.NodeIds = []primitive.ObjectID{t.NodeId}
		opts.Mode = constants.RunTypeSelectedNodes
	}

	// user
	if u := GetUserFromContext(c); u != nil {
		opts.UserId = u.Id
	}

	// run
	adminSvc := admin.GetSpiderAdminService()
	taskIds, err := adminSvc.Schedule(t.SpiderId, opts)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	HandleSuccessWithData(c, taskIds)
}

func PostTaskCancel(c *gin.Context) {
	type Payload struct {
		Force bool `json:"force,omitempty"`
	}

	// id
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// payload
	var p Payload
	if err := c.ShouldBindJSON(&p); err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// task
	t, err := service.NewModelService[models.Task]().GetById(id)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// validate
	if !utils.IsCancellable(t.Status) {
		HandleErrorInternalServerError(c, errors.New("task is not cancellable"))
		return
	}

	u := GetUserFromContext(c)

	// cancel
	schedulerSvc := scheduler.GetTaskSchedulerService()
	err = schedulerSvc.Cancel(id, u.Id, p.Force)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	HandleSuccess(c)
}

func GetTaskLogs(c *gin.Context) {
	// id
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// pagination
	p, err := GetPagination(c)
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	// logs
	logDriver := log.GetFileLogDriver()
	logs, err := logDriver.Find(id.Hex(), "", (p.Page-1)*p.Size, p.Size)
	if err != nil {
		if strings.HasSuffix(err.Error(), "Status:404 Not Found") {
			HandleSuccess(c)
			return
		}
		HandleErrorInternalServerError(c, err)
		return
	}
	total, err := logDriver.Count(id.Hex(), "")
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	HandleSuccessWithListData(c, logs, total)
}
