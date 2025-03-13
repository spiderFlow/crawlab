package controllers

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"sync"

	"github.com/crawlab-team/crawlab/core/fs"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/gin-gonic/gin"
)

func GetBaseFileListDir(rootPath, path string) (response *Response[[]interfaces.FsFileInfo], err error) {
	fsSvc, err := fs.GetBaseFileFsSvc(rootPath)
	if err != nil {
		return GetErrorResponse[[]interfaces.FsFileInfo](err)
	}

	files, err := fsSvc.List(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return GetErrorResponse[[]interfaces.FsFileInfo](err)
		}
	}

	return GetDataResponse(files)
}

func GetBaseFileContent(rootPath, path string) (response *Response[string], err error) {
	fsSvc, err := fs.GetBaseFileFsSvc(rootPath)
	if err != nil {
		return GetErrorResponse[string](err)
	}

	data, err := fsSvc.GetFile(path)
	if err != nil {
		return GetErrorResponse[string](err)
	}

	return GetDataResponse(string(data))
}

func GetBaseFileInfo(rootPath, path string) (response *Response[interfaces.FsFileInfo], err error) {
	fsSvc, err := fs.GetBaseFileFsSvc(rootPath)
	if err != nil {
		return GetErrorResponse[interfaces.FsFileInfo](err)
	}

	info, err := fsSvc.GetFileInfo(path)
	if err != nil {
		return GetErrorResponse[interfaces.FsFileInfo](err)
	}

	return GetDataResponse(info)
}

type PostBaseFileSaveOneParams struct {
	Path string `json:"path" form:"path"`
	Data string `json:"data"`
}

func PostBaseFileSaveOne(rootPath, path, data string) (response *Response[any], err error) {
	fsSvc, err := fs.GetBaseFileFsSvc(rootPath)
	if err != nil {
		return GetErrorResponse[any](err)
	}

	if err := fsSvc.Save(path, []byte(data)); err != nil {
		return GetErrorResponse[any](err)
	}

	return GetDataResponse(any(data))
}

func PostBaseFileSaveOneForm(rootPath, path string, file *multipart.FileHeader) (response *Response[any], err error) {
	fsSvc, err := fs.GetBaseFileFsSvc(rootPath)
	if err != nil {
		return GetErrorResponse[any](err)
	}

	f, err := file.Open()
	if err != nil {
		return GetErrorResponse[any](err)
	}

	fileData, err := io.ReadAll(f)
	if err != nil {
		return GetErrorResponse[any](err)
	}

	if err := fsSvc.Save(path, fileData); err != nil {
		return GetErrorResponse[any](err)
	}

	return GetDataResponse[any](nil)
}

func PostBaseFileSaveMany(rootPath string, form *multipart.Form) (response *Response[any], err error) {
	fsSvc, err := fs.GetBaseFileFsSvc(rootPath)
	if err != nil {
		return GetErrorResponse[any](err)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(form.File))
	for path := range form.File {
		go func(path string) {
			file := form.File[path][0]
			if err != nil {
				logger.Warnf("invalid file header: %s", path)
				logger.Error(err.Error())
				wg.Done()
				return
			}
			f, err := file.Open()
			if err != nil {
				logger.Warnf("unable to open file: %s", path)
				logger.Error(err.Error())
				wg.Done()
				return
			}
			fileData, err := io.ReadAll(f)
			if err != nil {
				logger.Warnf("unable to read file: %s", path)
				logger.Error(err.Error())
				wg.Done()
				return
			}
			if err := fsSvc.Save(path, fileData); err != nil {
				logger.Warnf("unable to save file: %s", path)
				logger.Error(err.Error())
				wg.Done()
				return
			}
			wg.Done()
		}(path)
	}
	wg.Wait()

	return GetDataResponse[any](nil)
}

func PostBaseFileSaveDir(rootPath, path string) (response *Response[any], err error) {
	fsSvc, err := fs.GetBaseFileFsSvc(rootPath)
	if err != nil {
		return GetErrorResponse[any](err)
	}

	if err := fsSvc.CreateDir(path); err != nil {
		return GetErrorResponse[any](err)
	}

	return GetDataResponse[any](nil)
}

func PostBaseFileRename(rootPath, path, newPath string) (response *Response[any], err error) {
	fsSvc, err := fs.GetBaseFileFsSvc(rootPath)
	if err != nil {
		return GetErrorResponse[any](err)
	}

	if err := fsSvc.Rename(path, newPath); err != nil {
		return GetErrorResponse[any](err)
	}

	return GetDataResponse[any](nil)
}

func DeleteBaseFile(rootPath, path string) (response *Response[any], err error) {
	if path == "~" {
		path = "."
	}

	fsSvc, err := fs.GetBaseFileFsSvc(rootPath)
	if err != nil {
		return GetErrorResponse[any](err)
	}

	if err := fsSvc.Delete(path); err != nil {
		return GetErrorResponse[any](err)
	}
	_, err = fsSvc.GetFileInfo(".")
	if err != nil {
		_ = fsSvc.CreateDir("/")
	}

	return GetDataResponse[any](nil)
}

func PostBaseFileCopy(rootPath, path, newPath string) (response *Response[any], err error) {
	fsSvc, err := fs.GetBaseFileFsSvc(rootPath)
	if err != nil {
		return GetErrorResponse[any](err)
	}

	if err := fsSvc.Copy(path, newPath); err != nil {
		return GetErrorResponse[any](err)
	}

	return GetDataResponse[any](nil)
}

func PostBaseFileExport(rootPath string, c *gin.Context) (err error) {
	fsSvc, err := fs.GetBaseFileFsSvc(rootPath)
	if err != nil {
		return err
	}

	// zip file path
	zipFilePath, err := fsSvc.Export()
	if err != nil {
		return err
	}

	// download
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFilePath))
	c.File(zipFilePath)

	return nil
}
