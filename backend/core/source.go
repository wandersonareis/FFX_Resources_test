package core

import (
	"ffxresources/backend/common"
	"ffxresources/backend/models"
	"io/fs"
	"os"
	"path/filepath"
)

type Source struct {
	Path       string
	Info       os.FileInfo
	Name       string
	NamePrefix string
	Type       models.NodeType
	Size       int64
	Extension  string
	EntryPath  string
	Parent     string
	IsDir      bool
}

func NewSource(path string) (*Source, error) {
	cPath := filepath.Clean(path)
	info, err := os.Stat(cPath)
	if err != nil {
		return nil, err
	}

	var size int64
	if !info.IsDir() {
		size = info.Size()
	}

	source := &Source{
		Path:       cPath,
		Info:       info,
		Size:       size,
		Name:       info.Name(),
		NamePrefix: common.RemoveFileExtension(info.Name()),
		Type:       guessFileType(cPath),
		Extension:  filepath.Ext(cPath),
		EntryPath:  filepath.Join(cPath, info.Name()),
		Parent:     filepath.Dir(cPath),
		IsDir:      info.IsDir(),
	}

	return source, nil

}

func (s *Source) ReadDir() ([]fs.DirEntry, error) {
	if s.IsDir {
		return os.ReadDir(s.Path)
	}

	return os.ReadDir(s.Parent)
}
