package controllers

import (
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/fs"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	mongo2 "github.com/crawlab-team/crawlab/core/mongo"
	"github.com/crawlab-team/crawlab/core/spider"
	"github.com/crawlab-team/crawlab/core/spider/admin"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetSpiderById handles getting a spider by ID
func GetSpiderById(_ *gin.Context, params *GetByIdParams) (response *Response[models.Spider], err error) {
	id, err := primitive.ObjectIDFromHex(params.Id)
	if err != nil {
		return GetErrorResponse[models.Spider](errors.BadRequestf("invalid id format"))
	}
	s, err := service.NewModelService[models.Spider]().GetById(id)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return GetErrorResponse[models.Spider](errors.NotFoundf("spider not found"))
	}
	if err != nil {
		return GetErrorResponse[models.Spider](err)
	}

	// stat
	s.Stat, err = service.NewModelService[models.SpiderStat]().GetById(s.Id)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return GetErrorResponse[models.Spider](err)
		}
	}

	// data collection (compatible to old version)
	if s.ColName == "" && !s.ColId.IsZero() {
		col, err := service.NewModelService[models.DataCollection]().GetById(s.ColId)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return GetErrorResponse[models.Spider](err)
			}
		} else {
			s.ColName = col.Name
		}
	}

	// git
	if utils.IsPro() && !s.GitId.IsZero() {
		s.Git, err = service.NewModelService[models.Git]().GetById(s.GitId)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return GetErrorResponse[models.Spider](err)
			}
		}
	}

	return GetDataResponse(*s)
}

// GetSpiderList handles getting a list of spiders with optional stats
func GetSpiderList(c *gin.Context, params *GetListParams) (response *ListResponse[models.Spider], err error) {
	// get all list
	all := params.All
	if all {
		return NewController[models.Spider]().GetAll(params)
	}

	// get list
	withStats := c.Query("stats")
	if withStats == "" {
		return NewController[models.Spider]().GetList(c, params)
	}

	// get list with stats
	return getSpiderListWithStats(params)
}

func getSpiderListWithStats(params *GetListParams) (response *ListResponse[models.Spider], err error) {
	query, err := GetFilterQueryFromListParams(params)
	if err != nil {
		return GetErrorListResponse[models.Spider](errors.BadRequestf("invalid request parameters: %v", err))
	}
	// get list
	spiders, err := service.NewModelService[models.Spider]().GetMany(query, &mongo2.FindOptions{
		Sort:  params.Sort,
		Skip:  params.Size * (params.Page - 1),
		Limit: params.Size,
	})
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return GetErrorListResponse[models.Spider](err)
		}
		return GetListResponse[models.Spider]([]models.Spider{}, 0)
	}
	if len(spiders) == 0 {
		return GetListResponse[models.Spider]([]models.Spider{}, 0)
	}

	// ids
	var ids []primitive.ObjectID
	var gitIds []primitive.ObjectID
	for _, s := range spiders {
		ids = append(ids, s.Id)
		if !s.GitId.IsZero() {
			gitIds = append(gitIds, s.GitId)
		}
	}

	// total count
	total, err := service.NewModelService[models.Spider]().Count(query)
	if err != nil {
		return GetErrorListResponse[models.Spider](err)
	}

	// stat list
	spiderStats, err := service.NewModelService[models.SpiderStat]().GetMany(bson.M{"_id": bson.M{"$in": ids}}, nil)
	if err != nil {
		return GetErrorListResponse[models.Spider](err)
	}

	// cache stat list to dict
	dict := map[primitive.ObjectID]models.SpiderStat{}
	var taskIds []primitive.ObjectID
	for _, st := range spiderStats {
		if st.Tasks > 0 {
			taskCount := int64(st.Tasks)
			st.AverageWaitDuration = int64(math.Round(float64(st.WaitDuration) / float64(taskCount)))
			st.AverageRuntimeDuration = int64(math.Round(float64(st.RuntimeDuration) / float64(taskCount)))
			st.AverageTotalDuration = int64(math.Round(float64(st.TotalDuration) / float64(taskCount)))
		}
		dict[st.Id] = st

		if !st.LastTaskId.IsZero() {
			taskIds = append(taskIds, st.LastTaskId)
		}
	}

	// task list and stats
	var tasks []models.Task
	dictTask := map[primitive.ObjectID]models.Task{}
	dictTaskStat := map[primitive.ObjectID]models.TaskStat{}
	if len(taskIds) > 0 {
		// task list
		queryTask := bson.M{
			"_id": bson.M{
				"$in": taskIds,
			},
		}
		tasks, err = service.NewModelService[models.Task]().GetMany(queryTask, nil)
		if err != nil {
			return GetErrorListResponse[models.Spider](err)
		}

		// task stats list
		taskStats, err := service.NewModelService[models.TaskStat]().GetMany(queryTask, nil)
		if err != nil {
			return GetErrorListResponse[models.Spider](err)
		}

		// cache task stats to dict
		for _, st := range taskStats {
			dictTaskStat[st.Id] = st
		}

		// cache task list to dict
		for _, t := range tasks {
			st, ok := dictTaskStat[t.Id]
			if ok {
				t.Stat = &st
			}
			dictTask[t.SpiderId] = t
		}
	}

	// git list
	var gits []models.Git
	if len(gitIds) > 0 && utils.IsPro() {
		gits, err = service.NewModelService[models.Git]().GetMany(bson.M{"_id": bson.M{"$in": gitIds}}, nil)
		if err != nil {
			return GetErrorListResponse[models.Spider](err)
		}
	}

	// cache git list to dict
	dictGit := map[primitive.ObjectID]models.Git{}
	for _, g := range gits {
		dictGit[g.Id] = g
	}

	// iterate list again
	var data []models.Spider
	for _, s := range spiders {
		// spider stat
		st, ok := dict[s.Id]
		if ok {
			s.Stat = &st

			// last task
			t, ok := dictTask[s.Id]
			if ok {
				s.Stat.LastTask = &t
			}
		}

		// git
		if !s.GitId.IsZero() && utils.IsPro() {
			g, ok := dictGit[s.GitId]
			if ok {
				s.Git = &g
			}
		}

		// add to list
		data = append(data, s)
	}

	// response
	return GetListResponse(data, total)
}

// PostSpider handles creating a new spider
func PostSpider(c *gin.Context, params *PostParams[models.Spider]) (response *Response[models.Spider], err error) {
	s := params.Data

	if s.Mode == "" {
		s.Mode = constants.RunTypeRandom
	}
	if s.Priority == 0 {
		s.Priority = 5
	}

	// user
	u := GetUserFromContext(c)

	// add
	s.SetCreated(u.Id)
	s.SetUpdated(u.Id)
	id, err := service.NewModelService[models.Spider]().InsertOne(s)
	if err != nil {
		return GetErrorResponse[models.Spider](err)
	}
	s.SetId(id)

	// add stat
	st := models.SpiderStat{}
	st.SetId(id)
	st.SetCreated(u.Id)
	st.SetUpdated(u.Id)
	_, err = service.NewModelService[models.SpiderStat]().InsertOne(st)
	if err != nil {
		return GetErrorResponse[models.Spider](err)
	}

	// create folder
	fsSvc, err := getSpiderFsSvcById(id)
	if err != nil {
		return GetErrorResponse[models.Spider](err)
	}
	err = fsSvc.CreateDir(".")
	if err != nil {
		return GetErrorResponse[models.Spider](err)
	}

	// create template if available
	if utils.IsPro() && s.Template != "" {
		if templateSvc := spider.GetSpiderTemplateRegistryService(); templateSvc != nil {
			err = templateSvc.CreateTemplate(s.Id)
			if err != nil {
				return GetErrorResponse[models.Spider](err)
			}
		}
	}

	return GetDataResponse(s)
}

// PutSpiderById handles updating a spider by ID
func PutSpiderById(c *gin.Context, params *PutByIdParams[models.Spider]) (response *Response[models.Spider], err error) {
	id, err := primitive.ObjectIDFromHex(params.Id)
	if err != nil {
		return GetErrorResponse[models.Spider](errors.BadRequestf("invalid id format"))
	}

	u := GetUserFromContext(c)
	modelSvc := service.NewModelService[models.Spider]()

	params.Data.SetUpdated(u.Id)
	if params.Data.Id.IsZero() {
		params.Data.SetId(id)
	}

	err = modelSvc.ReplaceById(id, params.Data)
	if err != nil {
		return GetErrorResponse[models.Spider](err)
	}

	s, err := modelSvc.GetById(id)
	if err != nil {
		return GetErrorResponse[models.Spider](err)
	}

	return GetDataResponse(*s)
}

// DeleteSpiderById handles deleting a spider by ID
func DeleteSpiderById(_ *gin.Context, params *DeleteByIdParams) (response *Response[models.Spider], err error) {
	id, err := primitive.ObjectIDFromHex(params.Id)
	if err != nil {
		return GetErrorResponse[models.Spider](errors.BadRequestf("invalid id format"))
	}

	s, err := service.NewModelService[models.Spider]().GetById(id)
	if err != nil {
		return GetErrorResponse[models.Spider](errors.NotFoundf("spider not found"))
	}

	if err := mongo2.RunTransaction(func(context mongo.SessionContext) (err error) {
		// delete spider
		err = service.NewModelService[models.Spider]().DeleteById(id)
		if err != nil {
			return err
		}

		// delete spider stat
		err = service.NewModelService[models.SpiderStat]().DeleteById(id)
		if err != nil {
			return err
		}

		// related tasks
		tasks, err := service.NewModelService[models.Task]().GetMany(bson.M{"spider_id": id}, nil)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			return nil
		}

		// task ids
		var taskIds []primitive.ObjectID
		for _, t := range tasks {
			taskIds = append(taskIds, t.Id)
		}

		// delete related tasks
		err = service.NewModelService[models.Task]().DeleteMany(bson.M{"_id": bson.M{"$in": taskIds}})
		if err != nil {
			return err
		}

		// delete related task stats
		err = service.NewModelService[models.TaskStat]().DeleteMany(bson.M{"_id": bson.M{"$in": taskIds}})
		if err != nil {
			return err
		}

		// delete tasks logs
		wg := sync.WaitGroup{}
		wg.Add(len(taskIds))
		for _, id := range taskIds {
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

		return nil
	}); err != nil {
		return GetErrorResponse[models.Spider](err)
	}

	if !s.GitId.IsZero() {
		go func() {
			// delete spider directory
			fsSvc, err := getSpiderFsSvcById(s.Id)
			if err != nil {
				logger.Errorf("failed to get spider fs service: %v", err)
				return
			}
			err = fsSvc.Delete(".")
			if err != nil {
				logger.Errorf("failed to delete spider directory: %v", err)
				return
			}
		}()
	}

	return GetDataResponse(models.Spider{})
}

type DeleteSpiderListParams struct {
	Ids []string `json:"ids" validate:"required"`
}

// DeleteSpiderList handles deleting multiple spiders
func DeleteSpiderList(_ *gin.Context, params *DeleteSpiderListParams) (response *Response[models.Spider], err error) {
	var ids []primitive.ObjectID
	for _, id := range params.Ids {
		_id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return GetErrorResponse[models.Spider](errors.BadRequestf("invalid id format"))
		}
		ids = append(ids, _id)
	}

	// Fetch spiders before deletion
	spiders, err := service.NewModelService[models.Spider]().GetMany(bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}, nil)
	if err != nil {
		return nil, err
	}

	if err := mongo2.RunTransaction(func(context mongo.SessionContext) (err error) {
		// delete spiders
		if err := service.NewModelService[models.Spider]().DeleteMany(bson.M{
			"_id": bson.M{
				"$in": ids,
			},
		}); err != nil {
			return err
		}

		// delete spider stats
		if err := service.NewModelService[models.SpiderStat]().DeleteMany(bson.M{
			"_id": bson.M{
				"$in": ids,
			},
		}); err != nil {
			return err
		}

		// related tasks
		tasks, err := service.NewModelService[models.Task]().GetMany(bson.M{"spider_id": bson.M{"$in": ids}}, nil)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			return nil
		}

		// task ids
		var taskIds []primitive.ObjectID
		for _, t := range tasks {
			taskIds = append(taskIds, t.Id)
		}

		// delete related tasks
		if err := service.NewModelService[models.Task]().DeleteMany(bson.M{"_id": bson.M{"$in": taskIds}}); err != nil {
			return err
		}

		// delete related task stats
		if err := service.NewModelService[models.TaskStat]().DeleteMany(bson.M{"_id": bson.M{"$in": taskIds}}); err != nil {
			return err
		}

		// delete tasks logs
		wg := sync.WaitGroup{}
		wg.Add(len(taskIds))
		for _, id := range taskIds {
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

		return nil
	}); err != nil {
		return GetErrorResponse[models.Spider](err)
	}

	// Delete spider directories
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(spiders))
		for i := range spiders {
			go func(s *models.Spider) {
				defer wg.Done()

				// Skip spider with git
				if !s.GitId.IsZero() {
					return
				}

				// Delete spider directory
				fsSvc, err := getSpiderFsSvcById(s.Id)
				if err != nil {
					logger.Errorf("failed to get spider fs service: %v", err)
					return
				}
				err = fsSvc.Delete(".")
				if err != nil {
					logger.Errorf("failed to delete spider directory: %v", err)
					return
				}
			}(&spiders[i])
		}
		wg.Wait()
	}()

	return GetDataResponse(models.Spider{})
}

func GetSpiderListDir(c *gin.Context, params *GetBaseFileListDirParams) (response *Response[[]interfaces.FsFileInfo], err error) {
	rootPath, err := getSpiderRootPathByContext(c)
	if err != nil {
		HandleErrorForbidden(c, err)
		return
	}
	return GetBaseFileListDir(rootPath, params)
}

func GetSpiderFile(c *gin.Context, params *GetBaseFileFileParams) (response *Response[string], err error) {
	rootPath, err := getSpiderRootPathByContext(c)
	if err != nil {
		HandleErrorForbidden(c, err)
		return
	}
	return GetBaseFileFile(rootPath, params)
}

func GetSpiderFileInfo(c *gin.Context) {
	rootPath, err := getSpiderRootPathByContext(c)
	if err != nil {
		HandleErrorForbidden(c, err)
		return
	}
	GetBaseFileFileInfo(rootPath, c)
}

func PostSpiderSaveFile(c *gin.Context) {
	rootPath, err := getSpiderRootPathByContext(c)
	if err != nil {
		HandleErrorForbidden(c, err)
		return
	}
	PostBaseFileSaveFile(rootPath, c)
}

func PostSpiderSaveFiles(c *gin.Context) {
	rootPath, err := getSpiderRootPathByContext(c)
	if err != nil {
		HandleErrorForbidden(c, err)
		return
	}
	targetDirectory := c.PostForm("targetDirectory")
	PostBaseFileSaveFiles(filepath.Join(rootPath, targetDirectory), c)
}

func PostSpiderSaveDir(c *gin.Context) {
	rootPath, err := getSpiderRootPathByContext(c)
	if err != nil {
		HandleErrorForbidden(c, err)
		return
	}
	PostBaseFileSaveDir(rootPath, c)
}

func PostSpiderRenameFile(c *gin.Context) {
	rootPath, err := getSpiderRootPathByContext(c)
	if err != nil {
		HandleErrorForbidden(c, err)
		return
	}
	PostBaseFileRenameFile(rootPath, c)
}

func DeleteSpiderFile(c *gin.Context) {
	rootPath, err := getSpiderRootPathByContext(c)
	if err != nil {
		HandleErrorForbidden(c, err)
		return
	}
	DeleteBaseFileFile(rootPath, c)
}

func PostSpiderCopyFile(c *gin.Context) {
	rootPath, err := getSpiderRootPathByContext(c)
	if err != nil {
		HandleErrorForbidden(c, err)
		return
	}
	PostBaseFileCopyFile(rootPath, c)
}

func PostSpiderExport(c *gin.Context) {
	rootPath, err := getSpiderRootPathByContext(c)
	if err != nil {
		HandleErrorForbidden(c, err)
		return
	}
	PostBaseFileExport(rootPath, c)
}

func PostSpiderRun(c *gin.Context) (response *Response[[]primitive.ObjectID], err error) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return GetErrorResponse[[]primitive.ObjectID](errors.BadRequestf("invalid id format"))
	}

	// options
	var opts interfaces.SpiderRunOptions
	if err := c.ShouldBindJSON(&opts); err != nil {
		return GetErrorResponse[[]primitive.ObjectID](err)
	}

	// user
	if u := GetUserFromContext(c); u != nil {
		opts.UserId = u.GetId()
	}

	// schedule tasks
	taskIds, err := admin.GetSpiderAdminService().Schedule(id, &opts)
	if err != nil {
		return GetErrorResponse[[]primitive.ObjectID](err)
	}

	HandleSuccessWithData(c, taskIds)
	return GetDataResponse(taskIds)
}

func GetSpiderResults(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		HandleErrorBadRequest(c, err)
		return
	}

	s, err := service.NewModelService[models.Spider]().GetById(id)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// params
	pagination := MustGetPagination(c)
	query := getResultListQuery(c)

	col := mongo2.GetMongoCol(s.ColName)

	var results []bson.M
	err = col.Find(mongo2.GetMongoQuery(query), mongo2.GetMongoOpts(&mongo2.ListOptions{
		Sort:  []mongo2.ListSort{{"_id", mongo2.SortDirectionDesc}},
		Skip:  pagination.Size * (pagination.Page - 1),
		Limit: pagination.Size,
	})).All(&results)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	total, err := mongo2.GetMongoCol(s.ColName).Count(mongo2.GetMongoQuery(query))
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	HandleSuccessWithListData(c, results, total)
}

func getSpiderFsSvc(s *models.Spider) (svc interfaces.FsService, err error) {
	workspacePath := utils.GetWorkspace()
	fsSvc := fs.NewFsService(filepath.Join(workspacePath, s.Id.Hex()))

	return fsSvc, nil
}

func getSpiderFsSvcById(id primitive.ObjectID) (svc interfaces.FsService, err error) {
	s, err := service.NewModelService[models.Spider]().GetById(id)
	if err != nil {
		logger.Errorf("failed to get spider: %v", err)
		return nil, err
	}
	return getSpiderFsSvc(s)
}

func getSpiderRootPathByContext(c *gin.Context) (rootPath string, err error) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return "", err
	}
	s, err := service.NewModelService[models.Spider]().GetById(id)
	if err != nil {
		return "", err
	}
	return utils.GetSpiderRootPath(s)
}
