package controllers

import (
	"github.com/crawlab-team/fizz/openapi"
	"github.com/gin-gonic/gin"
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
