package core

import (
	"ffxresources/backend/common"
	"ffxresources/backend/models"
	"io/fs"
	"os"
	"path/filepath"
)

type SpiraFileInfo struct {
	Name         string          `json:"name"`
	NamePrefix   string          `json:"name_prefix"`
	Type         models.NodeType `json:"type"`
	Size         int64           `json:"size"`
	Extension    string          `json:"extension"`
	EntryPath    string          `json:"entry_path"`
	Parent       string          `json:"parent"`
	IsDir        bool            `json:"is_dir"`
	ClonedItems  []string        `json:"cloned_items"`
	Path         string          `json:"path"`
	RelativePath string          `json:"relative_path"`
}

func NewSpiraFileInfo(path string, gamePart models.GameVersion) (*SpiraFileInfo, error) {
	cPath := filepath.Clean(path)
	info, err := os.Stat(cPath)
	if err != nil {
		return nil, err
	}

	var size int64
	if info != nil && !info.IsDir() {
		size = info.Size()
	}

	source := &SpiraFileInfo{
		Path: cPath,
		Size:         size,
		Name:         common.RecursiveRemoveFileExtension(info.Name()),
		NamePrefix:   common.RemoveOneFileExtension(info.Name()),
		Type:         guessFileType(cPath),
		Extension:    filepath.Ext(cPath),
		EntryPath:    filepath.Join(cPath, info.Name()),
		Parent:       filepath.Dir(cPath),
		IsDir:        info.IsDir(),
		RelativePath: common.GetRelativePathFromMarker(cPath),
	}

	return source, nil

}

func (s *SpiraFileInfo) ReadDir() ([]fs.DirEntry, error) {
	if s.IsDir {
		return os.ReadDir(s.Path)
	}

	return os.ReadDir(s.Parent)
}
