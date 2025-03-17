package controllers

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/entity"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/mongo"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/trace"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var logger = utils.NewLogger("Controllers")

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

func GetFilterQueryFromListParams(params *GetListParams) (q bson.M, err error) {
	if params.Conditions == "" {
		return nil, nil
	}
	conditions, err := GetFilterFromConditionString(params.Conditions)
	if err != nil {
		return nil, err
	}
	return utils.FilterToQuery(conditions)
}

func GetFilterQueryFromConditionString(condStr string) (q bson.M, err error) {
	if condStr == "" {
		return nil, nil
	}
	conditions, err := GetFilterFromConditionString(condStr)
	if err != nil {
		return nil, err
	}
	return utils.FilterToQuery(conditions)
}

func GetFilterFromConditionString(condStr string) (f *entity.Filter, err error) {
	if condStr == "" {
		return nil, nil
	}
	var conditions []*entity.Condition
	if err := json.Unmarshal([]byte(condStr), &conditions); err != nil {
		return nil, err
	}

	for i, cond := range conditions {
		v := reflect.ValueOf(cond.Value)
		switch v.Kind() {
		case reflect.String:
			item := cond.Value.(string)
			// attempt to convert object id
			id, err := primitive.ObjectIDFromHex(item)
			if err == nil {
				conditions[i].Value = id
			} else {
				conditions[i].Value = item
			}
		case reflect.Float64:
			// JSON numbers are decoded as float64 by default
			switch cond.Value.(type) {
			case float64:
				num := cond.Value.(float64)
				// Check if it's a whole number
				if num == float64(int64(num)) {
					conditions[i].Value = int64(num)
				} else {
					conditions[i].Value = num
				}
			case int:
				num := cond.Value.(int)
				conditions[i].Value = int64(num)
			case int64:
				num := cond.Value.(int64)
				conditions[i].Value = num
			}
		case reflect.Bool:
			conditions[i].Value = cond.Value.(bool)
		case reflect.Slice, reflect.Array:
			var items []interface{}
			for i := 0; i < v.Len(); i++ {
				vItem := v.Index(i)
				item := vItem.Interface()

				switch typedItem := item.(type) {
				case string:
					// Try to convert to ObjectID first
					if id, err := primitive.ObjectIDFromHex(typedItem); err == nil {
						items = append(items, id)
					} else {
						items = append(items, typedItem)
					}
				case float64:
					if typedItem == float64(int64(typedItem)) {
						items = append(items, int64(typedItem))
					} else {
						items = append(items, typedItem)
					}
				case bool:
					items = append(items, typedItem)
				default:
					items = append(items, item)
				}
			}
			conditions[i].Value = items
		default:
			conditions[i].Value = cond.Value
		}
	}

	return &entity.Filter{
		IsOr:       false,
		Conditions: conditions,
	}, nil
}

// GetFilter Get entity.Filter from gin.Context
func GetFilter(c *gin.Context) (f *entity.Filter, err error) {
	condStr := c.Query(constants.FilterQueryFieldConditions)
	return GetFilterFromConditionString(condStr)
}

// GetFilterQuery Get bson.M from gin.Context
func GetFilterQuery(c *gin.Context) (q bson.M, err error) {
	f, err := GetFilter(c)
	if err != nil {
		return nil, err
	}

	if f == nil {
		return nil, nil
	}

	// TODO: implement logic OR

	return utils.FilterToQuery(f)
}

func MustGetFilterQuery(c *gin.Context) (q bson.M) {
	q, err := GetFilterQuery(c)
	if err != nil {
		return nil
	}
	return q
}

func getResultListQuery(c *gin.Context) (q mongo.ListQuery) {
	f, err := GetFilter(c)
	if err != nil {
		return q
	}
	for _, cond := range f.Conditions {
		q = append(q, mongo.ListQueryCondition{
			Key:   cond.Key,
			Op:    cond.Op,
			Value: utils.NormalizeObjectId(cond.Value),
		})
	}
	return q
}

func GetDefaultPagination() (p *entity.Pagination) {
	return &entity.Pagination{
		Page: constants.PaginationDefaultPage,
		Size: constants.PaginationDefaultSize,
	}
}

func GetPagination(c *gin.Context) (p *entity.Pagination, err error) {
	var _p entity.Pagination
	if err := c.ShouldBindQuery(&_p); err != nil {
		return GetDefaultPagination(), err
	}
	if _p.Page == 0 {
		_p.Page = constants.PaginationDefaultPage
	}
	if _p.Size == 0 {
		_p.Size = constants.PaginationDefaultSize
	}
	return &_p, nil
}

func MustGetPagination(c *gin.Context) (p *entity.Pagination) {
	p, err := GetPagination(c)
	if err != nil || p == nil {
		return GetDefaultPagination()
	}
	return p
}

func GetSortsFromString(sortStr string) (sorts []entity.Sort, err error) {
	if sortStr == "" {
		return nil, nil
	}
	if err := json.Unmarshal([]byte(sortStr), &sorts); err != nil {
		return nil, err
	}
	return sorts, nil
}

func GetSortOptionFromString(sortStr string) (sort bson.D, err error) {
	sorts, err := GetSortsFromString(sortStr)
	if err != nil {
		return nil, err
	}
	if sorts == nil || len(sorts) == 0 {
		return bson.D{{"_id", -1}}, nil
	}
	return SortsToOption(sorts)
}

// GetSorts Get entity.Sort from gin.Context
func GetSorts(c *gin.Context) (sorts []entity.Sort, err error) {
	// bind
	sortStr := c.Query(constants.SortQueryField)
	if err := json.Unmarshal([]byte(sortStr), &sorts); err != nil {
		return nil, err
	}
	return sorts, nil
}

// GetSortsOption Get entity.Sort from gin.Context
func GetSortsOption(c *gin.Context) (sort bson.D, err error) {
	sorts, err := GetSorts(c)
	if err != nil {
		return nil, err
	}

	if sorts == nil || len(sorts) == 0 {
		return bson.D{{"_id", -1}}, nil
	}

	return SortsToOption(sorts)
}

func MustGetSortOption(c *gin.Context) (sort bson.D) {
	sort, err := GetSortsOption(c)
	if err != nil {
		return nil
	}
	return sort
}

// SortsToOption Translate entity.Sort to bson.D
func SortsToOption(sorts []entity.Sort) (sort bson.D, err error) {
	sort = bson.D{}
	for _, s := range sorts {
		switch s.Direction {
		case constants.ASCENDING:
			sort = append(sort, bson.E{Key: s.Key, Value: 1})
		case constants.DESCENDING:
			sort = append(sort, bson.E{Key: s.Key, Value: -1})
		}
	}
	if len(sort) == 0 {
		sort = bson.D{{"_id", -1}}
	}
	return sort, nil
}

type Response[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Error   string `json:"error"`
}

type ListResponse[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Total   int    `json:"total"`
	Data    []T    `json:"data"`
	Error   string `json:"error"`
}

type VoidResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func GetDataResponse[T any](model T) (res *Response[T], err error) {
	return &Response[T]{
		Status:  constants.HttpResponseStatusOk,
		Message: constants.HttpResponseMessageSuccess,
		Data:    model,
	}, nil
}

func GetListResponse[T any](models []T, total int) (res *ListResponse[T], err error) {
	return &ListResponse[T]{
		Status:  constants.HttpResponseStatusOk,
		Message: constants.HttpResponseMessageSuccess,
		Data:    models,
		Total:   total,
	}, nil
}

func GetVoidResponse() (res *VoidResponse, err error) {
	return &VoidResponse{
		Status:  constants.HttpResponseStatusOk,
		Message: constants.HttpResponseMessageSuccess,
	}, nil
}

func GetErrorResponse[T any](err error) (res *Response[T], err2 error) {
	return &Response[T]{
		Status:  constants.HttpResponseStatusOk,
		Message: constants.HttpResponseMessageError,
		Error:   err.Error(),
	}, err
}

func GetErrorVoidResponse(err error) (res *VoidResponse, err2 error) {
	return &VoidResponse{
		Status:  constants.HttpResponseStatusOk,
		Message: constants.HttpResponseMessageError,
		Error:   err.Error(),
	}, err
}

func GetErrorListResponse[T any](err error) (res *ListResponse[T], err2 error) {
	return &ListResponse[T]{
		Status:  constants.HttpResponseStatusOk,
		Message: constants.HttpResponseMessageError,
		Error:   err.Error(),
	}, err
}

func handleError(statusCode int, c *gin.Context, err error) {
	if utils.IsDev() {
		trace.PrintError(err)
	}
	c.AbortWithStatusJSON(statusCode, entity.Response{
		Status:  constants.HttpResponseStatusOk,
		Message: constants.HttpResponseMessageError,
		Error:   err.Error(),
	})
}

func HandleError(statusCode int, c *gin.Context, err error) {
	handleError(statusCode, c, err)
}

func HandleErrorBadRequest(c *gin.Context, err error) {
	HandleError(http.StatusBadRequest, c, err)
}

func HandleErrorForbidden(c *gin.Context, err error) {
	HandleError(http.StatusForbidden, c, err)
}

func HandleErrorUnauthorized(c *gin.Context, err error) {
	HandleError(http.StatusUnauthorized, c, err)
}

func HandleErrorNotFound(c *gin.Context, err error) {
	HandleError(http.StatusNotFound, c, err)
}

func HandleErrorInternalServerError(c *gin.Context, err error) {
	HandleError(http.StatusInternalServerError, c, err)
}

func HandleSuccess(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusOK, entity.Response{
		Status:  constants.HttpResponseStatusOk,
		Message: constants.HttpResponseMessageSuccess,
	})
}

func HandleSuccessWithData(c *gin.Context, data interface{}) {
	c.AbortWithStatusJSON(http.StatusOK, entity.Response{
		Status:  constants.HttpResponseStatusOk,
		Message: constants.HttpResponseMessageSuccess,
		Data:    data,
	})
}

func HandleSuccessWithListData(c *gin.Context, data interface{}, total int) {
	c.AbortWithStatusJSON(http.StatusOK, entity.ListResponse{
		Status:  constants.HttpResponseStatusOk,
		Message: constants.HttpResponseMessageSuccess,
		Data:    data,
		Total:   total,
	})
}
