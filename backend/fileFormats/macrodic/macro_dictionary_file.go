package macrodic

import (
	"ffxresources/backend/datastore"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
)

// IMacroDictionaryFile é o contrato de um arquivo de dicionário de macros
// (uma localization). Embute interfaces.IBinaryFile para o ciclo de vida
// universal (extração/importação/salvamento via interface) e adiciona os
// métodos próprios do formato: mapa indexado por macro ID e publicação no
// datastore, que agrega as localizations como um único dicionário lógico.
type IMacroDictionaryFile interface {
	interfaces.IBinaryFile[datastore.IGlobalLocalizedMacroStringObject]
	// GetMapObjects retorna o mapa já formado desta localization
	// (chave chunk*0x100+índice), compatível com datastore.MacroObjects.
	GetMapObjects() datastore.MacroObjects
	// PublishStrings mescla esta localization no datastore versionado.
	PublishStrings() error
	GetLocalization() string
	GetVersion() models.GameVersion
}

var (
	_ IMacroDictionaryFile                                                         = (*MacroDictionaryBinaryFile)(nil)
	_ interfaces.IBinaryFile[datastore.IGlobalLocalizedMacroStringObject]          = (*MacroDictionaryBinaryFile)(nil)
)
