package controllers

import (
	"net/http"

	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/entity"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/trace"
	"github.com/gin-gonic/gin"
)

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

func GetSuccessDataResponse[T any](model T) (res *Response[T], err error) {
	return &Response[T]{
		Status:  constants.HttpResponseStatusOk,
		Message: constants.HttpResponseMessageSuccess,
		Data:    model,
	}, nil
}

func GetSuccessListResponse[T any](models []T, total int) (res *ListResponse[T], err error) {
	return &ListResponse[T]{
		Status:  constants.HttpResponseStatusOk,
		Message: constants.HttpResponseMessageSuccess,
		Data:    models,
		Total:   total,
	}, nil
}

func GetErrorDataResponse[T any](err error) (res *Response[T], err2 error) {
	return &Response[T]{
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
