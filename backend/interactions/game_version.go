package interactions

import (
	"ffxresources/backend/common"
)

// Este arquivo centraliza a leitura da versão ativa do jogo.
//
// A versão é common.GameVersion ("ffx"/"ffx2"/"lastmiss"), persistida
// como string no config.json e espelhada em common.CurrentGameVersion().
// Não existem mais variantes int (1/2), sufixos v1/v2 nem env GAME_VERSION.

// GameVersion retorna a versão ativa.
func (i *InteractionService) GameVersion() common.GameVersion {
	if i == nil || i.ffxAppConfig == nil {
		return common.CurrentGameVersion()
	}
	return i.ffxAppConfig.GetGameVersion()
}

// CurrentGameVersion é um atalho de pacote para a versão ativa.
func CurrentGameVersion() common.GameVersion {
	return NewInteractionService().GameVersion()
}
