package fs

import (
	"github.com/crawlab-team/crawlab/core/entity"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/trace"
	"github.com/google/uuid"
	"io"
	"os"
	"path/filepath"
)

type Service struct {
	// settings
	rootPath  string
	skipNames []string
}

func (svc *Service) List(path string) (files []interfaces.FsFileInfo, err error) {
	// Normalize the provided path
	normPath := filepath.Clean(path)
	if normPath == "." {
		normPath = ""
	}
	fullPath := filepath.Join(svc.rootPath, normPath)

	// Temporary map to hold directory information and their children
	dirMap := make(map[string]*entity.FsFileInfo)

	// Use filepath.Walk to recursively traverse directories
	err = filepath.Walk(fullPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(svc.rootPath, p)
		if err != nil {
			return err
		}

		fi := &entity.FsFileInfo{
			Name:      info.Name(),
			Path:      filepath.ToSlash(relPath),
			FullPath:  p,
			Extension: filepath.Ext(p),
			IsDir:     info.IsDir(),
			FileSize:  info.Size(),
			ModTime:   info.ModTime(),
			Mode:      info.Mode(),
			Children:  nil,
		}

		// Skip files/folders matching the pattern
		for _, name := range svc.skipNames {
			if fi.Name == name {
				return nil
			}
		}

		if info.IsDir() {
			dirMap[p] = fi
		}

		if parentDir := filepath.Dir(p); parentDir != p && dirMap[parentDir] != nil {
			dirMap[parentDir].Children = append(dirMap[parentDir].Children, fi)
		}

		return nil
	})

	if rootInfo, ok := dirMap[fullPath]; ok {
		for _, info := range rootInfo.GetChildren() {
			files = append(files, info)
		}
	}

	return files, err
}

func (svc *Service) GetFile(path string) (data []byte, err error) {
	return os.ReadFile(filepath.Join(svc.rootPath, path))
}

func (svc *Service) GetFileInfo(path string) (file interfaces.FsFileInfo, err error) {
	f, err := os.Stat(filepath.Join(svc.rootPath, path))
	if err != nil {
		return nil, err
	}
	return &entity.FsFileInfo{
		Name:      f.Name(),
		Path:      path,
		FullPath:  filepath.Join(svc.rootPath, path),
		Extension: filepath.Ext(path),
		IsDir:     f.IsDir(),
		FileSize:  f.Size(),
		ModTime:   f.ModTime(),
		Mode:      f.Mode(),
		Children:  nil,
	}, nil
}

func (svc *Service) Save(path string, data []byte) (err error) {
	// Create directories if not exist
	dir := filepath.Dir(filepath.Join(svc.rootPath, path))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Write file
	return os.WriteFile(filepath.Join(svc.rootPath, path), data, 0644)
}

func (svc *Service) CreateDir(path string) (err error) {
	return os.MkdirAll(filepath.Join(svc.rootPath, path), 0755)
}

func (svc *Service) Rename(path, newPath string) (err error) {
	oldPath := filepath.Join(svc.rootPath, path)
	newFullPath := filepath.Join(svc.rootPath, newPath)
	return os.Rename(oldPath, newFullPath)
}

func (svc *Service) Delete(path string) (err error) {
	fullPath := filepath.Join(svc.rootPath, path)
	return os.RemoveAll(fullPath)
}

func (svc *Service) Copy(path, newPath string) (err error) {
	srcPath := filepath.Join(svc.rootPath, path)
	destPath := filepath.Join(svc.rootPath, newPath)

	// Get source info
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	// If source is file, copy it
	if !srcInfo.IsDir() {
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)

		return err
	} else {
		// If source is directory, copy it recursively
		return utils.CopyDir(srcPath, destPath)
	}
}

func (svc *Service) Export() (resultPath string, err error) {
	zipFilePath := filepath.Join(os.TempDir(), uuid.New().String()+".zip")
	if err := utils.ZipDirectory(svc.rootPath, zipFilePath); err != nil {
		return "", trace.TraceError(err)
	}

	return zipFilePath, nil
}

func NewFsService(path string) (svc interfaces.FsService) {
	return &Service{
		rootPath:  path,
		skipNames: []string{".git"},
	}
}
