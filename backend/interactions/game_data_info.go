package interactions

import (
	"ffxresources/backend/core"
	"os"
	"strings"
)

type GameDataInfo struct {
	GameData          core.GameFiles    `json:"game_data"`
	ExtractLocation   ExtractLocation   `json:"extract_location"`
	TranslateLocation TranslateLocation `json:"translate_location"`
	ImportLocation    ImportLocation    `json:"import_location"`
}

func NewGameDataInfo(path string) *GameDataInfo {
	gameData, err := core.NewGameData(path)
	if err != nil {
		return nil
	}

	return &GameDataInfo{
		GameData:          *gameData,
		ExtractLocation:   *NewExtractLocation(),
		TranslateLocation: *NewTranslateLocation(),
		ImportLocation:    *NewImportLocation(),
	}
}

func (g *GameDataInfo) InitializeLocations(formatter ITextFormatter) {
	g.ExtractLocation.GenerateTargetOutput(formatter, g)
	g.TranslateLocation.GenerateTargetOutput(formatter, g)
	g.ImportLocation.GenerateTargetOutput(formatter, g)
}

// CreateRelativePath sets the RelativeGameDataPath field of the GameDataInfo struct
// by computing the relative path from the FullFilePath to the target directory.
// If a target directory is provided as an argument, it will be used; otherwise,
// the default target directory from GameLocation will be used.
//
// Parameters:
//
//	target (optional) - A variadic string parameter that specifies the target directory.
//
// Example:
//
//	s.CreateRelativePath("C:\\Games\\TargetDir")
func (g *GameDataInfo) CreateRelativePath(target ...string) {
	targetPath := NewInteraction().GameLocation.TargetDirectory

	if len(target) > 0 {
		targetPath = target[0]
	}

	if strings.HasPrefix(g.GameData.FullFilePath, targetPath) {
		g.GameData.RelativeGameDataPath = strings.TrimPrefix(g.GameData.FullFilePath, targetPath+string(os.PathSeparator))
	}
}
