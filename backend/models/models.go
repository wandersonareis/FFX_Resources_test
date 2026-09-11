package models

import (
	
	
)

type NodeType int

const (
	None NodeType = iota
	Drive
	File
	Folder
	Dialogs
	DialogsSpecial
	Tutorial
	Kernel
	Dcp
	DcpParts
	Lockit
	LockitParts
)

type GameVersion int

const (
	FFX GameVersion = iota + 1
	FFX2
)

func (gv GameVersion) String() string {
	switch gv {
	case FFX:
		return "FFX"
	case FFX2:
		return "FFX-2"
	default:
		return "Unknown"
	}
}

// GameVersionModel é a identidade textual da versão do jogo ("ffx"/"ffx2").
//
// Existe para diferenciar da GameVersion numérica (int 1/2) usada como chave
// de mapas versionados e em binários. Use Model() para converter e
// ParseGameVersionModel/NewGameVersionModelFromInt para o caminho inverso.
type GameVersionModel string

const (
	GameVersionModelFFX  GameVersionModel = "ffx"
	GameVersionModelFFX2 GameVersionModel = "ffx2"
)

// GameVersionString é um alias legível para GameVersionModel, mantido para
// chamadas que esperam explicitamente uma string de versão.
type GameVersionString = GameVersionModel

// Model retorna a identidade textual ("ffx"/"ffx2") da versão numérica.
// Valores desconhecidos caem para "ffx".
func (gv GameVersion) Model() GameVersionModel {
	if gv == FFX2 {
		return GameVersionModelFFX2
	}
	return GameVersionModelFFX
}

// GameVersionString retorna "ffx"/"ffx2" (forma string de Model()).
func (gv GameVersion) GameVersionString() string {
	return string(gv.Model())
}

// Number converte o modelo textual de volta para a versão numérica.
// Valores desconhecidos caem para FFX.
func (m GameVersionModel) Number() int {
	switch GameVersionModel(normalizeGameVersionModelString(string(m))) {
	case GameVersionModelFFX2:
		return int(FFX2)
	default:
		return int(FFX)
	}
}

// ToGameVersion converte o modelo textual para GameVersion (int).
func (m GameVersionModel) ToGameVersion() GameVersion {
	return GameVersion(m.Number())
}

func normalizeGameVersionModelString(s string) string {
	switch s {
	case "ffx2", "FFX2", "FFX-2", "ffx-2", "v2", "2":
		return string(GameVersionModelFFX2)
	default:
		return string(GameVersionModelFFX)
	}
}

// NewGameVersionModelFromInt cria um GameVersionModel a partir de 1/2.
// Valores fora do intervalo caem para "ffx".
func NewGameVersionModelFromInt(v int) GameVersionModel {
	return GameVersion(v).Model()
}

// NewGameVersionModelFromString cria um GameVersionModel a partir de
// "ffx"/"ffx2" (aceita variações como "FFX-2", "v2", "2").
func NewGameVersionModelFromString(s string) GameVersionModel {
	return GameVersionModel(normalizeGameVersionModelString(s))
}

// ParseGameVersionModel é um alias de NewGameVersionModelFromString.
func ParseGameVersionModel(s string) GameVersionModel {
	return NewGameVersionModelFromString(s)
}

type (
	GameDataInfo struct {
		FilePath       string `json:"file_path"`
		ExtractedFile  string `json:"extracted_file"`
		TranslatedFile string `json:"translated_file"`
		ImportedFile   string `json:"imported_file"`
	}

	Pointer struct {
		Offset int64
		Value  uint32
	}

	FileComparisonEntry struct {
		FromFile string
		ToFile   string
	}
)
