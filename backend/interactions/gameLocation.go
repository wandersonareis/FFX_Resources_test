package interactions

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interfaces"
	"fmt"
)

type IGameLocation interface {
	interfaces.IInteractionBase

	IsSpira() error
}

type GameLocation struct {
	*interactionBase
}

func newGameLocation() IGameLocation {
	defaultDirName := "data"
	return &GameLocation{
		interactionBase: &interactionBase{
			defaultDirName: defaultDirName,
		},
	}
}

func (g *GameLocation) GetTargetDirectory() string {
	path, _ := g.interactionBase.GetTargetDirectoryBase(ConfigGameFilesLocation)
	return path.(string)
}

func (g *GameLocation) SetTargetDirectory(path string) error {
	return g.interactionBase.SetTargetDirectoryBase(ConfigGameFilesLocation, path)
}

func (g *GameLocation) ProvideTargetDirectory() error {
	path := g.GetTargetDirectory()

	err := g.interactionBase.ProviderTargetDirectoryBase(ConfigGameFilesLocation, path)
	if err != nil {
		return err
	}

	return nil
}

func (g *GameLocation) IsSpira() error {
	version, err := common.CheckFFXPath(g.GetTargetDirectory())
	if err != nil {
		return err
	}

	if version > 0 {
		return nil
	}

	return fmt.Errorf("path does not contain a valid spira file: %s", g.GetTargetDirectory())
}
