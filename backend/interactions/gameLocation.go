package interactions

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interfaces"
)

type IGameLocation interface {
	interfaces.IInteractionBase
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
