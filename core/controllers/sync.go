package controllers

import (
	"path/filepath"

	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/gin-gonic/gin"
)

func GetSyncScan(c *gin.Context) {
	workspacePath := utils.GetWorkspace()
	dirPath := filepath.Join(workspacePath, c.Param("id"), c.Param("path"))
	files, err := utils.ScanDirectory(dirPath)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	HandleSuccessWithData(c, files)
}

func GetSyncDownload(c *gin.Context) {
	workspacePath := utils.GetWorkspace()
	filePath := filepath.Join(workspacePath, c.Param("id"), c.Param("path"))
	c.File(filePath)
}
