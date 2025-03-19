package controllers

import (
	"net/http"
	"time"

	"github.com/loopfz/gadgeto/tonic"

	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/mongo"
	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
)

func init() {
	tonic.SetErrorHook(func(context *gin.Context, err error) (int, interface{}) {
		response := gin.H{
			"error": errors.Unwrap(err).Error(),
		}
		status := http.StatusInternalServerError
		constErr, ok := errors.AsType[errors.ConstError](err)
		if ok {
			switch {
			case errors.Is(constErr, errors.NotFound):
				status = http.StatusNotFound
			case errors.Is(constErr, errors.BadRequest):
				status = http.StatusBadRequest
			case errors.Is(constErr, errors.Unauthorized):
				status = http.StatusUnauthorized
			case errors.Is(constErr, errors.Forbidden):
				status = http.StatusForbidden
			default:
				status = http.StatusInternalServerError
			}
		} else {
			status = http.StatusInternalServerError
		}
		return status, response
	})
}

type Action struct {
	Method      string
	Path        string
	Name        string
	Description string
	HandlerFunc interface{}
}

type BaseController[T any] struct {
	modelSvc *service.ModelService[T]
	actions  []Action
}

// GetListParams represents parameters for GetList with pagination
type GetListParams struct {
	Conditions string `query:"conditions" description:"Filter conditions. Format: [{\"key\":\"name\",\"op\":\"eq\",\"value\":\"test\"}]"`
	Sort       string `query:"sort" description:"Sort options"`
	Page       int    `query:"page" default:"1" description:"Page number" minimum:"1"`
	Size       int    `query:"size" default:"10" description:"Page size" minimum:"1"`
	All        bool   `query:"all" default:"false" description:"Whether to get all items"`
}

func (ctr *BaseController[T]) GetList(_ *gin.Context, params *GetListParams) (response *ListResponse[T], err error) {
	// get all if query field "all" is set true
	if params.All {
		return ctr.GetAll(params)
	}

	return ctr.GetWithPagination(params)
}

type GetByIdParams struct {
	Id string `path:"id" description:"The ID of the item to get" format:"objectid" pattern:"^[0-9a-fA-F]{24}$"`
}

func (ctr *BaseController[T]) GetById(_ *gin.Context, params *GetByIdParams) (response *Response[T], err error) {
	id, err := primitive.ObjectIDFromHex(params.Id)
	if err != nil {
		return GetErrorResponse[T](errors.BadRequestf("invalid id format"))
	}

	model, err := ctr.modelSvc.GetById(id)
	if err != nil {
		return nil, err
	}

	return GetDataResponse(*model)
}

type PostParams[T any] struct {
	Data T `json:"data" description:"The data to create" validate:"required"`
}

func (ctr *BaseController[T]) Post(c *gin.Context, params *PostParams[T]) (response *Response[T], err error) {
	u := GetUserFromContext(c)
	m := any(&params.Data).(interfaces.Model)
	m.SetId(primitive.NewObjectID())
	m.SetCreated(u.Id)
	m.SetUpdated(u.Id)
	col := ctr.modelSvc.GetCol()
	res, err := col.GetCollection().InsertOne(col.GetContext(), m)
	if err != nil {
		return GetErrorResponse[T](err)
	}

	result, err := ctr.modelSvc.GetById(res.InsertedID.(primitive.ObjectID))
	if err != nil {
		return GetErrorResponse[T](err)
	}

	return GetDataResponse(*result)
}

type PutByIdParams[T any] struct {
	Id   string `path:"id" description:"The ID of the item to update" format:"objectid" pattern:"^[0-9a-fA-F]{24}$"`
	Data T      `json:"data" description:"The data to update" validate:"required"`
}

func (ctr *BaseController[T]) PutById(c *gin.Context, params *PutByIdParams[T]) (response *Response[T], err error) {
	id, err := primitive.ObjectIDFromHex(params.Id)
	if err != nil {
		return GetErrorResponse[T](errors.BadRequestf("invalid id format: %v", err))
	}

	u := GetUserFromContext(c)
	m := any(&params.Data).(interfaces.Model)
	m.SetUpdated(u.Id)
	if m.GetId().IsZero() {
		m.SetId(id)
	}

	if err := ctr.modelSvc.ReplaceById(id, params.Data); err != nil {
		return GetErrorResponse[T](err)
	}

	result, err := ctr.modelSvc.GetById(id)
	if err != nil {
		return GetErrorResponse[T](err)
	}

	return GetDataResponse(*result)
}

type PatchParams struct {
	Ids    []string `json:"ids" description:"The IDs of the items to update" validate:"required" items.type:"string" items.format:"objectid" items.pattern:"^[0-9a-fA-F]{24}$"`
	Update bson.M   `json:"update" description:"The update object" validate:"required"`
}

func (ctr *BaseController[T]) PatchList(c *gin.Context, params *PatchParams) (res *Response[T], err error) {
	var ids []primitive.ObjectID
	for _, id := range params.Ids {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return GetErrorResponse[T](errors.BadRequestf("invalid id format: %v", err))
		}
		ids = append(ids, objectId)
	}

	// Get user from context for updated_by
	u := GetUserFromContext(c)

	// query
	query := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}

	// Add updated_by and updated_ts to the update object
	updateObj := params.Update
	updateObj["updated_by"] = u.Id
	updateObj["updated_ts"] = time.Now()

	// update
	if err := ctr.modelSvc.UpdateMany(query, bson.M{"$set": updateObj}); err != nil {
		return GetErrorResponse[T](err)
	}

	// Return an empty response with success status
	var emptyModel T
	return GetDataResponse(emptyModel)
}

type DeleteByIdParams struct {
	Id string `path:"id" description:"The ID of the item to delete" format:"objectid" pattern:"^[0-9a-fA-F]{24}$"`
}

func (ctr *BaseController[T]) DeleteById(c *gin.Context, params *DeleteByIdParams) (res *Response[T], err error) {
	params.Id = c.Param("id")
	id, err := primitive.ObjectIDFromHex(params.Id)
	if err != nil {
		return GetErrorResponse[T](errors.BadRequestf("invalid id format: %v", err))
	}

	if err := ctr.modelSvc.DeleteById(id); err != nil {
		return GetErrorResponse[T](err)
	}

	var emptyModel T
	return GetDataResponse(emptyModel)
}

type DeleteListParams struct {
	Ids []string `json:"ids" description:"The IDs of the items to delete" items.type:"string" items.format:"objectid" items.pattern:"^[0-9a-fA-F]{24}$"`
}

func (ctr *BaseController[T]) DeleteList(_ *gin.Context, params *DeleteListParams) (res *Response[T], err error) {
	var ids []primitive.ObjectID
	for _, id := range params.Ids {
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return GetErrorResponse[T](err)
		}
		ids = append(ids, objectId)
	}

	if err := ctr.modelSvc.DeleteMany(bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}); err != nil {
		return GetErrorResponse[T](err)
	}

	var emptyModel T
	return GetDataResponse(emptyModel)
}

// GetAll retrieves all items based on filter and sort
func (ctr *BaseController[T]) GetAll(params *GetListParams) (response *ListResponse[T], err error) {
	// Get filter query
	query, err := GetFilterQueryFromListParams(params)
	if err != nil {
		return GetErrorListResponse[T](errors.BadRequestf("invalid request parameters: %v", err))
	}

	// Get sort options
	sort, err := GetSortOptionFromString(params.Sort)
	if err != nil {
		return GetErrorListResponse[T](errors.BadRequestf("invalid sort format: %v", err))
	}

	// Get models
	models, err := ctr.modelSvc.GetMany(query, &mongo.FindOptions{
		Sort: sort,
	})
	if err != nil {
		return nil, err
	}

	// Total count
	total, err := ctr.modelSvc.Count(query)
	if err != nil {
		return nil, err
	}

	// Response
	return GetListResponse(models, total)
}

// GetWithPagination retrieves items with pagination
func (ctr *BaseController[T]) GetWithPagination(params *GetListParams) (response *ListResponse[T], err error) {
	// Get filter query
	query, err := GetFilterQueryFromListParams(params)
	if err != nil {
		return GetErrorListResponse[T](errors.BadRequestf("invalid request parameters: %v", err))
	}

	// Get sort options
	sort, err := GetSortOptionFromString(params.Sort)
	if err != nil {
		return GetErrorListResponse[T](errors.BadRequestf("invalid sort format: %v", err))
	}

	// Get models
	models, err := ctr.modelSvc.GetMany(query, &mongo.FindOptions{
		Sort:  sort,
		Skip:  params.Size * (params.Page - 1),
		Limit: params.Size,
	})
	if err != nil {
		if errors.Is(err, mongo2.ErrNoDocuments) {
			return GetListResponse[T](nil, 0)
		} else {
			return nil, err
		}
	}

	// Total count
	total, err := ctr.modelSvc.Count(query)
	if err != nil {
		return nil, err
	}

	// Response
	return GetListResponse(models, total)
}

func NewController[T any](actions ...Action) *BaseController[T] {
	ctr := &BaseController[T]{
		modelSvc: service.NewModelService[T](),
		actions:  actions,
	}
	return ctr
}
