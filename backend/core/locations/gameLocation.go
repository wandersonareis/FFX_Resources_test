package locations

import (
	"ffxresources/backend/common"
)

type IGameLocation interface {
	GetTargetDirectory() string
	SetTargetDirectory(path string)
	IsSpira() error
}

type GameLocation struct {
	LocationBase
}

const defaultDirName = "data"

func NewGameLocation() IGameLocation {
	return &GameLocation{
		LocationBase: NewLocationBase(defaultDirName),
	}
}

func (g GameLocation) IsSpira() error {
	return common.ContainsNewUSPCPath(g.TargetDirectory)
}
