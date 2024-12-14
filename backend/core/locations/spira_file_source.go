package locations

import (
	"ffxresources/backend/core"
	"ffxresources/backend/interfaces"
	"path/filepath"
)

type Source struct {
	FileInfo *core.SpiraFileInfo `json:"file_info"`
}

func NewSource(path string, gamePart core.GamePart) (interfaces.ISource, error) {
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

/* func (g *Source) updateGameDataFromSource(source *SpiraFileInfo, gamePart GamePart) {
	source.ClonedItems = g.GetGamePartDuplicates(gamePart)
} */

func (g *Source) Get() *core.SpiraFileInfo {
	return g.FileInfo
}

func (g *Source) Set(fileInfo *core.SpiraFileInfo) {
	g.FileInfo = fileInfo
}

func (g *Source) GetGamePartDuplicates(namePrefix string, gamePart core.GamePart) []string {
	switch gamePart {
	case core.FFX:
		//TODO: return NewFfxDuplicate().AddFfxTextDuplicate()
		fallthrough
	case core.FFX2:
		dupes := core.NewFfx2Duplicate()
		dupes.AddFfx2TextDuplicate()
		return dupes.TryFind(namePrefix)
	}

	return nil
}
