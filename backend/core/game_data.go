package core

import (
	"ffxresources/backend/common"
	"ffxresources/backend/models"
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

func NewGameData(path string) (*GameFiles, error) {
	data := &GameFiles{}
	source, err := NewSource(path)
	if err != nil {
		return data, err
	}

	data.updateGameDataFromSource(source)

	return data, nil
}

func (g *GameFiles) updateGameDataFromSource(source *Source) {
	g.Name = source.Name
	g.NamePrefix = source.NamePrefix
	g.Size = source.Size
	g.Type = source.Type
	g.IsDir = source.IsDir
	g.Parent = source.Parent
	g.Extension = source.Extension
	g.FullFilePath = source.Path
	g.RelativeGameDataPath = common.GetRelativePathFromMarker(g.FullFilePath)
	g.ClonedItems = NewFfx2Duplicate().TryFind(source.NamePrefix)
}
