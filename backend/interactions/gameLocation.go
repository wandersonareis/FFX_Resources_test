package interactions

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interfaces"
	"fmt"
)

type IGameLocation interface {
	interfaces.IInteractionBase

	WithTargetDirectory(path string) IGameLocation
	IsSpira() error
}

type GameLocation struct {
	*interactionBase
}

func newGameLocation(path string, config ISetAppConfig) IGameLocation {
	gameLocation := &GameLocation{
		interactionBase: &interactionBase{
			targetDir: path,
			config:    config,
			configKey: "GameFilesLocation",
		},
	}
	gameLocation.SetTargetDirectory(path)
	return gameLocation
}

func (g *GameLocation) WithTargetDirectory(path string) IGameLocation {
	_ = g.SetTargetDirectory(path)
	return g
}

func (g *GameLocation) IsSpira() error {
	version, err := common.CheckFFXPath(g.GetTargetDirectory())
	if err != nil {
		return err
	}

	if version.IsValid() {
		return nil
	}

	return fmt.Errorf("path does not contain a valid spira file: %s", g.GetTargetDirectory())
}
