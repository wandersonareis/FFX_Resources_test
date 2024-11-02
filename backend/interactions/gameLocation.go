package interactions

import "ffxresources/backend/common"

type GameLocation struct {
	LocationBase
}

const defaultDirName = "data"

func NewGameLocation() *GameLocation {
	return &GameLocation{
		LocationBase: NewLocationBase(defaultDirName),
	}
}

func (g GameLocation) IsSpira() error {
	return common.ContainsNewUSPCPath(g.TargetDirectory)
}
