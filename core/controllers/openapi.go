package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wI2L/fizz/openapi"
)

func GetOpenAPI(c *gin.Context) {
	info := &openapi.Info{
		Title:       "Crawlab API",
		Description: "REST API for Crawlab",
		Version:     "0.7.0",
	}
	handleFunc := globalWrapper.GetFizz().OpenAPI(info, "json")
	handleFunc(c)
}
