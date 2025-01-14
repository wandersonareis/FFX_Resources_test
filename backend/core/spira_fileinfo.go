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

func NewSpiraFileInfo(path string) (*SpiraFileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	size := info.Size()
	if info.IsDir() {
		size, err = getDirSize(path)
		if err != nil {
			return nil, err
		}
	}

	source := &SpiraFileInfo{
		Path:         path,
		Size:         size,
		Name:         common.RecursiveRemoveFileExtension(info.Name()),
		NamePrefix:   common.RemoveOneFileExtension(info.Name()),
		Type:         guessFileType(path),
		Extension:    filepath.Ext(path),
		EntryPath:    path,
		Parent:       filepath.Dir(path),
		IsDir:        info.IsDir(),
		RelativePath: "",
	}

	if source.Type == models.Dcp || source.Type == models.Lockit {
		source.RelativePath = common.GetRelativePathFromMarker(path)
	}

	return source, nil
}

func (s *SpiraFileInfo) ReadDir() ([]fs.DirEntry, error) {
	if s.IsDir {
		return os.ReadDir(s.Path)
	}

	return os.ReadDir(s.Parent)
}
