package interactions

import "ffxresources/backend/common"

type IGameLocation interface {
	GetTargetDirectory() string
	SetTargetDirectory(path string)
	IsSpira() error
}

type GameLocation struct {
	InteractionBase
}

const defaultDirName = "data"

func NewGameLocation() IGameLocation {
	return &GameLocation{
		InteractionBase: newInteractionBase(defaultDirName),
	}
}

func (g GameLocation) IsSpira() error {
	return common.ContainsNewUSPCPath(g.TargetDirectory)
}
