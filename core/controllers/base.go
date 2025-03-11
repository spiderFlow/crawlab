package controllers

import (
	"github.com/crawlab-team/crawlab/core/entity"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/mongo"
	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
)

type Action struct {
	Method      string
	Path        string
	HandlerFunc interface{}
}

type BaseController[T any] struct {
	modelSvc *service.ModelService[T]
	actions  []Action
}

type GetByIdParams struct {
	Id string `path:"id" description:"The ID of the item to get"`
}

// GetAllParams represents parameters for GetAll method
type GetAllParams struct {
	Query bson.M `json:"query"`
	Sort  bson.D `json:"sort"`
}

// GetListParams represents parameters for GetList with pagination
type GetListParams struct {
	Query      bson.M             `json:"query"`
	Sort       bson.D             `json:"sort"`
	Pagination *entity.Pagination `json:"pagination"`
	All        bool               `query:"all" description:"Whether to get all items"`
}

func (ctr *BaseController[T]) GetById(_ *gin.Context, params *GetByIdParams) (response *Response[T], err error) {
	id, err := primitive.ObjectIDFromHex(params.Id)
	if err != nil {
		return nil, errors.BadRequestf("invalid id: %s", params.Id)
	}

	model, err := ctr.modelSvc.GetById(id)
	if err != nil {
		return nil, err
	}

	return GetSuccessDataResponse(*model)
}

func (ctr *BaseController[T]) GetList(c *gin.Context, params *GetListParams) (response *ListResponse[T], err error) {
	// get all if query field "all" is set true
	all := params.All || MustGetFilterAll(c)

	// Prepare parameters
	query := MustGetFilterQuery(c)
	sort := MustGetSortOption(c)

	if all {
		allParams := &GetAllParams{
			Query: query,
			Sort:  sort,
		}
		return ctr.GetAll(c, allParams)
	}

	// get list with pagination
	pagination := MustGetPagination(c)
	listParams := &GetListParams{
		Query:      query,
		Sort:       sort,
		Pagination: pagination,
	}
	return ctr.GetWithPagination(c, listParams)
}

func (ctr *BaseController[T]) Post(c *gin.Context) (response *Response[T], err error) {
	var model T
	if err := c.ShouldBindJSON(&model); err != nil {
		return GetErrorDataResponse[T](err)
	}
	u := GetUserFromContext(c)
	m := any(&model).(interfaces.Model)
	m.SetId(primitive.NewObjectID())
	m.SetCreated(u.Id)
	m.SetUpdated(u.Id)
	col := ctr.modelSvc.GetCol()
	res, err := col.GetCollection().InsertOne(col.GetContext(), m)
	if err != nil {
		return GetErrorDataResponse[T](err)
	}

	result, err := ctr.modelSvc.GetById(res.InsertedID.(primitive.ObjectID))
	if err != nil {
		return GetErrorDataResponse[T](err)
	}

	return GetSuccessDataResponse(*result)
}

func (ctr *BaseController[T]) PutById(c *gin.Context) (response *Response[T], err error) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return GetErrorDataResponse[T](err)
	}

	var model T
	if err := c.ShouldBindJSON(&model); err != nil {
		return GetErrorDataResponse[T](err)
	}

	u := GetUserFromContext(c)
	m := any(&model).(interfaces.Model)
	m.SetUpdated(u.Id)

	if err := ctr.modelSvc.ReplaceById(id, model); err != nil {
		return GetErrorDataResponse[T](err)
	}

	result, err := ctr.modelSvc.GetById(id)
	if err != nil {
		return GetErrorDataResponse[T](err)
	}

	return GetSuccessDataResponse(*result)
}

func (ctr *BaseController[T]) PatchList(c *gin.Context) (res *Response[T], err error) {
	type Payload struct {
		Ids    []primitive.ObjectID `json:"ids"`
		Update bson.M               `json:"update"`
	}

	var payload Payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		return GetErrorDataResponse[T](err)
	}

	// query
	query := bson.M{
		"_id": bson.M{
			"$in": payload.Ids,
		},
	}

	// update
	if err := ctr.modelSvc.UpdateMany(query, bson.M{"$set": payload.Update}); err != nil {
		return GetErrorDataResponse[T](err)
	}

	// Return an empty response with success status
	var emptyModel T
	return GetSuccessDataResponse(emptyModel)
}

func (ctr *BaseController[T]) DeleteById(c *gin.Context) (res *Response[T], err error) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return GetErrorDataResponse[T](err)
	}

	if err := ctr.modelSvc.DeleteById(id); err != nil {
		return GetErrorDataResponse[T](err)
	}

	var emptyModel T
	return GetSuccessDataResponse(emptyModel)
}

func (ctr *BaseController[T]) DeleteList(c *gin.Context) (res *Response[T], err error) {
	type Payload struct {
		Ids []string `json:"ids"`
	}

	var payload Payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		return GetErrorDataResponse[T](err)
	}

	var ids []primitive.ObjectID
	for _, id := range payload.Ids {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return GetErrorDataResponse[T](err)
		}
		ids = append(ids, objectId)
	}

	if err := ctr.modelSvc.DeleteMany(bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}); err != nil {
		return GetErrorDataResponse[T](err)
	}

	var emptyModel T
	return GetSuccessDataResponse(emptyModel)
}

// GetAll retrieves all items based on filter and sort
func (ctr *BaseController[T]) GetAll(_ *gin.Context, params *GetAllParams) (response *ListResponse[T], err error) {
	query := params.Query
	sort := params.Sort
	if sort == nil {
		sort = bson.D{{"_id", -1}}
	}

	models, err := ctr.modelSvc.GetMany(query, &mongo.FindOptions{
		Sort: sort,
	})
	if err != nil {
		return nil, err
	}

	total, err := ctr.modelSvc.Count(query)
	if err != nil {
		return nil, err
	}

	return GetSuccessListResponse(models, total)
}

// GetWithPagination retrieves items with pagination
func (ctr *BaseController[T]) GetWithPagination(_ *gin.Context, params *GetListParams) (response *ListResponse[T], err error) {
	// params
	pagination := params.Pagination
	query := params.Query
	sort := params.Sort

	if pagination == nil {
		pagination = GetDefaultPagination()
	}

	// get list
	models, err := ctr.modelSvc.GetMany(query, &mongo.FindOptions{
		Sort:  sort,
		Skip:  pagination.Size * (pagination.Page - 1),
		Limit: pagination.Size,
	})
	if err != nil {
		if errors.Is(err, mongo2.ErrNoDocuments) {
			return GetSuccessListResponse[T](nil, 0)
		} else {
			return nil, err
		}
	}

	// total count
	total, err := ctr.modelSvc.Count(query)
	if err != nil {
		return nil, err
	}

	// response
	return GetSuccessListResponse(models, total)
}

// getAll is kept for backward compatibility
func (ctr *BaseController[T]) getAll(c *gin.Context) (response *ListResponse[T], err error) {
	query := MustGetFilterQuery(c)
	sort := MustGetSortOption(c)

	params := &GetAllParams{
		Query: query,
		Sort:  sort,
	}

	return ctr.GetAll(c, params)
}

// getList is kept for backward compatibility
func (ctr *BaseController[T]) getList(c *gin.Context) (response *ListResponse[T], err error) {
	// params
	pagination := MustGetPagination(c)
	query := MustGetFilterQuery(c)
	sort := MustGetSortOption(c)

	params := &GetListParams{
		Query:      query,
		Sort:       sort,
		Pagination: pagination,
	}

	return ctr.GetWithPagination(c, params)
}

func NewController[T any](actions ...Action) *BaseController[T] {
	ctr := &BaseController[T]{
		modelSvc: service.NewModelService[T](),
		actions:  actions,
	}
	return ctr
}
