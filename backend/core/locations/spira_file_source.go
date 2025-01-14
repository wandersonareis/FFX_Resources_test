package locations

import (
	"ffxresources/backend/core"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"path/filepath"
	"sync"
)

type Source struct {
	FileInfo *core.SpiraFileInfo `json:"file_info"`
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

	fileInfo, err := core.NewSpiraFileInfo(absPath)
	if err != nil {
		return nil, err
	}

	return &Source{
		FileInfo: fileInfo,
	}, nil
}

func (g *Source) Get() *core.SpiraFileInfo {
	return g.FileInfo
}

func (g *Source) Set(fileInfo *core.SpiraFileInfo) {
	g.FileInfo = fileInfo
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
