package fs

import (
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/utils"
	"path/filepath"
)

func GetBaseFileFsSvc(rootPath string) (svc interfaces.FsService, err error) {
	workspacePath := utils.GetWorkspace()
	fsSvc := NewFsService(filepath.Join(workspacePath, rootPath))

	return fsSvc, nil
}
