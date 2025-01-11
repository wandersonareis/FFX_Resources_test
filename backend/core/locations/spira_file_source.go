package locations

import (
	"ffxresources/backend/core"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"path/filepath"
)

type Source struct {
	FileInfo *core.SpiraFileInfo `json:"file_info"`
}

func NewSource(path string, gamePart models.GameVersion) (interfaces.ISource, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	fileInfo, err := core.NewSpiraFileInfo(absPath, gamePart)
	if err != nil {
		return nil, err
	}

	fileInfo.ClonedItems = new(Source).GetGamePartDuplicates(fileInfo.NamePrefix, gamePart)

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

func (g *Source) GetGamePartDuplicates(namePrefix string, gamePart models.GameVersion) []string {
	switch gamePart {
	case models.FFX:
		//TODO: return NewFfxDuplicate().AddFfxTextDuplicate()
		fallthrough
	case models.FFX2:
		duplicateHandler := core.NewFfx2Duplicate()
		duplicateHandler.AddFfx2TextDuplicate()
		return duplicateHandler.TryFind(namePrefix)
	}

	return nil
}
