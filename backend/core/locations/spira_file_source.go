package locations

import (
	"ffxresources/backend/core"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"path/filepath"
	"sync"
)

type Source struct {
	FileInfo *models.SpiraFileInfo `json:"file_info"`
}

var ffx2FileDuplicates = &sync.Pool{
	New: func() interface{} {
		ffx2Duplicates := core.NewFfx2Duplicate()
		ffx2Duplicates.AddFfx2TextDuplicate()
		return ffx2Duplicates
	},
}

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

func (g *Source) PopulateDuplicatesFiles(gameVersion models.GameVersion) {
	switch gameVersion {
	case models.FFX:
		//TODO: return NewFfxDuplicate().AddFfxTextDuplicate()
		fallthrough
	case models.FFX2:
		ffx2FilesDupe := ffx2FileDuplicates.Get().(*core.Ffx2Duplicate)
		defer ffx2FileDuplicates.Put(ffx2FilesDupe)

		g.FileInfo.ClonedItems = ffx2FilesDupe.TryFind(g.FileInfo.NamePrefix)
	}
}
