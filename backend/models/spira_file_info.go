package models

import (
	"ffxresources/backend/common"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type SpiraFileInfo struct {
	Name         string      `json:"name"`
	NamePrefix   string      `json:"name_prefix"`
	Extension    string      `json:"extension"`
	IsDir        bool        `json:"is_dir"`
	ClonedItems  []string    `json:"cloned_items"`
	Path         string      `json:"path"`
	Parent       string      `json:"parent"`
	RelativePath string      `json:"relative_path"`
	Size         int64       `json:"size"`
	Type         NodeType    `json:"type"`
	Version      GameVersion `json:"version"`
}

func NewSpiraFileInfo(path string) (*SpiraFileInfo, error) {
	aPath, err := resolvePath(path)
	if err != nil {
		return nil, err
	}

	info, err := getFileInfo(aPath)
	if err != nil {
		return nil, err
	}

	version, err := determineVersion(aPath)
	if err != nil {
		return nil, err
	}

	relativePath := getRelativePath(aPath, version)
	if !info.IsDir() {
		version, err = common.CheckFFXPath(aPath)
		if err != nil {
			return nil, err
		}
		relativePath, err = common.RelativePathFromMatch(aPath)
		if err != nil {
			return nil, err
		}
	}

	size, err := computeSize(path, info)
	if err != nil {
		return nil, err
	}

	return createSpiraFileInfo(info, path, relativePath, size, version), nil
}

func resolvePath(path string) (string, error) {
	return filepath.Abs(path)
}

func getFileInfo(aPath string) (os.FileInfo, error) {
	return os.Stat(aPath)
}

func determineVersion(aPath string) (int, error) {
	v := getVersionFromPrefix(aPath)
	if v == 0 {
		return 0, fmt.Errorf("invalid path: %s", aPath)
	}
	return v, nil
}

func computeSize(path string, info os.FileInfo) (int64, error) {
	if info.IsDir() {
		return getDirSize(path)
	}
	return info.Size(), nil
}

func createSpiraFileInfo(info os.FileInfo, path, relativePath string, size int64, version int) *SpiraFileInfo {
	return &SpiraFileInfo{
		Name:         common.RecursiveRemoveFileExtension(info.Name()),
		NamePrefix:   common.RemoveOneFileExtension(info.Name()),
		Extension:    filepath.Ext(path),
		IsDir:        info.IsDir(),
		Path:         path,
		Parent:       filepath.Dir(path),
		RelativePath: relativePath,
		Size:         size,
		Version:      GameVersion(version),
		Type:         guessFileType(path),
	}
}

func (s *SpiraFileInfo) SetPath(path string) {
	s.Path = path
}

func (s *SpiraFileInfo) SetRelativePath(relativePath string) {
	s.RelativePath = relativePath
}

func (s *SpiraFileInfo) ReadDir() ([]fs.DirEntry, error) {
	if s.IsDir {
		return os.ReadDir(s.Path)
	}

	return os.ReadDir(s.Parent)
}

func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func getVersionFromPrefix(path string) int {
	rxFFX2 := regexp.MustCompile(`(?i)FFX-2`)
	rxFFX := regexp.MustCompile(`(?i)FFX`)

	switch {
	case rxFFX2.MatchString(path):
		return 2
	case rxFFX.MatchString(path):
		return 1
	default:
		return 0
	}
}

func getRelativePath(path string, version int) string {
	switch version {
	case 1:
		idx := strings.Index(path, "FFX"+string(os.PathSeparator))
		if idx < 0 {
			return ""
		}
		idx += len("FFX") + 1
		if idx >= len(path) {
			return ""
		}
		return path[idx:]
	case 2:
		idx := strings.Index(path, "FFX-2"+string(os.PathSeparator))
		if idx < 0 {
			return ""
		}
		idx += len("FFX-2") + 1
		if idx >= len(path) {
			return ""
		}
		return path[idx:]
	default:
		return ""
	}
}
