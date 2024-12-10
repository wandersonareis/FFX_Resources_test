package core

import (
	"ffxresources/backend/common"
	"ffxresources/backend/models"
	"path/filepath"
)

type GameFiles struct {
	Name                 string          `json:"name"`
	NamePrefix           string          `json:"name_prefix"`
	Size                 int64           `json:"size"`
	Type                 models.NodeType `json:"type"`
	Extension            string          `json:"extension"`
	Parent               string          `json:"parent"`
	IsDir                bool            `json:"is_dir"`
	FullFilePath         string          `json:"full_path"`
	RelativeGameDataPath string          `json:"relative_path"`
	ClonedItems          []string        `json:"cloned_items"`
}

func NewGameData(path string, gamePart GamePart) (*GameFiles, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	data := &GameFiles{}
	source, err := NewSpiraFileInfo(absPath)
	if err != nil {
		return data, err
	}

	data.updateGameDataFromSource(source, gamePart)

	return data, nil
}

func (g *GameFiles) updateGameDataFromSource(source *SpiraFileInfo, gamePart GamePart) {
	g.Name = source.Name
	g.NamePrefix = source.NamePrefix
	g.Size = source.Size
	g.Type = source.Type
	g.IsDir = source.IsDir
	g.Parent = source.Parent
	g.Extension = source.Extension
	g.FullFilePath = source.Path
	g.RelativeGameDataPath = common.GetRelativePathFromMarker(g.FullFilePath)
	g.ClonedItems = g.GetGamePartDuplicates(gamePart)
}

func (g *GameFiles) GetGamePartDuplicates(gamePart GamePart) []string {
	switch gamePart {
	case FFX:
		//NewFfxDuplicate().AddFfxTextDuplicate()
		fallthrough
	case FFX2:
		dupes := NewFfx2Duplicate()
		dupes.AddFfx2TextDuplicate()
		return dupes.TryFind(g.NamePrefix)
	}

	return nil
}
