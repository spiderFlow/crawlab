package controllers

import (
	"path/filepath"

	"github.com/crawlab-team/crawlab/core/entity"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/gin-gonic/gin"
)

type GetSyncScanParams struct {
	Id   string `path:"id" validate:"required"`
	Path string `query:"path"`
}

func GetSyncScan(_ *gin.Context, params *GetSyncScanParams) (response *Response[map[string]entity.FsFileInfo], err error) {
	workspacePath := utils.GetWorkspace()
	dirPath := filepath.Join(workspacePath, params.Id, params.Path)
	files, err := utils.ScanDirectory(dirPath)
	if err != nil {
		return GetErrorResponse[map[string]entity.FsFileInfo](err)
	}
	return GetDataResponse(files)
}

type GetSyncDownloadParams struct {
	Id   string `path:"id" validate:"required"`
	Path string `query:"path" validate:"required"`
}

func GetSyncDownload(c *gin.Context, params *GetSyncDownloadParams) (err error) {
	workspacePath := utils.GetWorkspace()
	filePath := filepath.Join(workspacePath, params.Id, params.Path)
	c.File(filePath)
	return nil
}
