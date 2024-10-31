package core

import (
	"ffxresources/backend/common"
	"ffxresources/backend/models"
)

type GameData struct {
	Name         string          `json:"name"`
	NamePrefix   string          `json:"name_prefix"`
	Size         int64           `json:"size"`
	Type         models.NodeType `json:"type"`
	Extension    string          `json:"extension"`
	Parent       string          `json:"parent"`
	IsDir        bool            `json:"is_dir"`
	AbsolutePath string          `json:"absolute_path"`
	RelativePath string          `json:"relative_path"`
	Clones       []string        `json:"clones"`
}

func NewGameData(path string) (*GameData, error) {
	data := &GameData{}
	source, err := NewSource(path)
	if err != nil {
		return data, err
	}

	data.updateGameDataFromSource(source)

	return data, nil
}

func (g *GameData) updateGameDataFromSource(source *Source) {
	g.Name = source.Name
	g.NamePrefix = source.NamePrefix
	g.Size = source.Size
	g.Type = source.Type
	g.IsDir = source.IsDir
	g.Parent = source.Parent
	g.Extension = source.Extension
	g.AbsolutePath = source.FullPath
	g.RelativePath = common.GetRelativePathFromMarker(g.AbsolutePath)
	g.Clones = NewFfx2Duplicate().TryFind(source.NamePrefix)
}
