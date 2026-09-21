package interactions

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interfaces"
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

// SetTargetDirectory persiste imediatamente no config.json (via base),
// garante a existência do diretório e sincroniza common.GameFilesRoot
// para que BuildTree/leituras usem o novo gamefiles sem restart.
func (g *GameLocation) SetTargetDirectory(path string) error {
	if err := g.interactionBase.SetTargetDirectory(path); err != nil {
		return err
	}
	target := g.interactionBase.GetTargetDirectory()
	if target != "" {
		_ = common.EnsurePathExists(target)
		common.SetGameFilesRoot(target)
	}
	return nil
}

func (g *GameLocation) IsSpira() error {
	_, err := common.CheckFFXPath(g.GetTargetDirectory())
	if err != nil {
		return err
	}

	return nil
}
