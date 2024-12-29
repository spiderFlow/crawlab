package controllers

import (
	"github.com/crawlab-team/crawlab/core/entity"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/gin-gonic/gin"
)

func GetSystemInfo(c *gin.Context) {
	info := &entity.SystemInfo{
		Edition: utils.GetEdition(),
		Version: utils.GetVersion(),
	}
	HandleSuccessWithData(c, info)
}
