package interactions

import (
	"ffxresources/backend/models"
)

// Este arquivo centraliza a leitura da versão ativa do jogo.
//
// FFXGameVersion() (em interaction.go) foi mantido por compatibilidade e
// retorna o provider models.IGameVersionProvider. As funções abaixo são o
// caminho preferido quando só é preciso o valor atual, sem lidar com o
// provider:
//
//   - GameVersion()       -> models.GameVersion (int 1/2, chave dos mapas versionados)
//   - GameVersionModel()  -> models.GameVersionModel ("ffx"/"ffx2")
//   - GameVersionString() -> string ("ffx"/"ffx2")
//
// Nenhuma importa encoding/datastore, então não há ciclo:
// interactions -> models (models não importa ninguém).

// GameVersion retorna a versão numérica ativa (models.FFX / models.FFX2).
func (i *InteractionService) GameVersion() models.GameVersion {
	if i == nil || i.ffxAppConfig == nil {
		return models.FFX
	}
	return models.GameVersion(i.ffxAppConfig.GetGameVersion())
}

// GameVersionModel retorna a identidade textual ativa ("ffx"/"ffx2").
func (i *InteractionService) GameVersionModel() models.GameVersionModel {
	return i.GameVersion().Model()
}

// GameVersionString retorna "ffx"/"ffx2" para paths e sufixos de arquivo.
func (i *InteractionService) GameVersionString() string {
	return string(i.GameVersionModel())
}

// CurrentGameVersion é um atalho de pacote para a versão numérica ativa.
func CurrentGameVersion() models.GameVersion {
	return NewInteractionService().GameVersion()
}

// CurrentGameVersionModel é um atalho de pacote para "ffx"/"ffx2".
func CurrentGameVersionModel() models.GameVersionModel {
	return NewInteractionService().GameVersionModel()
}

// CurrentGameVersionString é um atalho de pacote para "ffx"/"ffx2" (string).
func CurrentGameVersionString() string {
	return NewInteractionService().GameVersionString()
}
