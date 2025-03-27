package core

import (
	"ffxresources/backend/common"
	"ffxresources/backend/models"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

	if source.Type != models.Lockit {
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

// CreateRelativePath sets the RelativeGameDataPath of the source to a path relative to the given gameLocationPath.
// If the FullFilePath of the source starts with the gameLocationPath, the gameLocationPath is trimmed from the FullFilePath
// and the result is assigned to RelativeGameDataPath.
//
// Parameters:
//   - gameLocationPath: The path for game original files to which the FullFilePath should be made relative.
func (s *SpiraFileInfo) CreateRelativePath(source *SpiraFileInfo, gameLocationPath string) {
	if strings.HasPrefix(source.Path, gameLocationPath) {
		source.RelativePath = strings.TrimPrefix(source.Path, gameLocationPath+string(os.PathSeparator))
	}
}