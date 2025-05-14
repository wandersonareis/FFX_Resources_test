package locations

import (
	"ffxresources/backend/duplicateFilesData"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"path/filepath"
	"sync"
)

type Source struct {
	FileInfo *models.SpiraFileInfo `json:"file_info"`
}

var (
	ffx2Once           sync.Once
	ffx2FileDuplicates *duplicateFilesData.Ffx2DuplicateFiles
)

func NewSource(path string) (interfaces.ISource, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	fileInfo, err := models.NewSpiraFileInfo(absPath)
	if err != nil {
		return nil, err
	}

	return &Source{
		FileInfo: fileInfo,
	}, nil
}

func (g *Source) Get() models.SpiraFileInfo {
	return *g.FileInfo
}

func (g *Source) GetName() string {
	return g.FileInfo.Name
}

func (g *Source) GetNameWithoutExtension() string {
	return g.FileInfo.NamePrefix
}

func (g *Source) GetExtension() string {
	return g.FileInfo.Extension
}

func (g *Source) GetParentPath() string {
	return g.FileInfo.Parent
}

func (g *Source) GetPath() string {
	return g.FileInfo.Path
}

func (g *Source) SetPath(path string) {
	g.FileInfo.Path = path
}

func (g *Source) GetRelativePath() string {
	return g.FileInfo.RelativePath
}

func (g *Source) SetRelativePath(relativePath string) {
	g.FileInfo.RelativePath = relativePath
}

func (g *Source) GetSize() int64 {
	return g.FileInfo.Size
}

func (g *Source) GetType() models.NodeType {
	return g.FileInfo.Type
}

func (g *Source) IsDir() bool {
	return g.FileInfo.IsDir
}

func (g *Source) PopulateDuplicatesFiles() {
	switch g.FileInfo.Version {
	case models.FFX:
		//TODO: return NewFfxDuplicate().AddFfxTextDuplicate()
		fallthrough
	case models.FFX2:
		ffx2FilesDupe := getFfx2FileDuplicates()
		g.FileInfo.ClonedItems = ffx2FilesDupe.TryFind(g.FileInfo.NamePrefix)
	}
}

func getFfx2FileDuplicates() *duplicateFilesData.Ffx2DuplicateFiles {
	ffx2Once.Do(func() {
		instance := duplicateFilesData.NewFfx2DuplicateFiles()
		instance.PopulateDuplicatesFiles()
		ffx2FileDuplicates = instance
	})
	return ffx2FileDuplicates
}
