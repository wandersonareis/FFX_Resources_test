package interactions

import "ffxresources/backend/common"

type IGameLocation interface {
	GetTargetDirectory() string
	SetTargetDirectory(path string)
	IsSpira() error
}

type GameLocation struct {
	*interactionBase
}

func newGameLocation(ffxAppConfig IFFXAppConfig) IGameLocation {
	defaultDirName := "data"
	return &GameLocation{
		interactionBase: &interactionBase{
			ffxAppConfig:      ffxAppConfig,
			defaultDirName: defaultDirName,
		},
	}
}

func (g *GameLocation) GetTargetDirectory() string {
	path, _ := g.interactionBase.GetTargetDirectoryBase(ConfigGameFilesLocation)
	return path.(string)
}

func (g *GameLocation) SetTargetDirectory(path string) {
	g.interactionBase.SetTargetDirectoryBase(ConfigGameFilesLocation, path)
}

func (g *GameLocation) IsSpira() error {
	return common.ContainsNewUSPCPath(g.GetTargetDirectory())
}
