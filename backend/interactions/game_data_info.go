package interactions

import (
	"ffxresources/backend/core"
	"ffxresources/backend/logger"
	"os"
	"strings"
)

type IGameDataInfo interface {
	GetGameDataInfo() *GameDataInfo
	InitializeLocations(formatter ITextFormatter)
	CreateRelativePath(target ...string)
	GetGameData() *core.GameFiles
	SetGameData(gameData *core.GameFiles)
	GetExtractLocation() *ExtractLocation
	GetTranslateLocation() *TranslateLocation
	GetImportLocation() *ImportLocation
}

type GameDataInfo struct {
	GameData          core.GameFiles    `json:"game_data"`
	ExtractLocation   ExtractLocation   `json:"extract_location"`
	TranslateLocation TranslateLocation `json:"translate_location"`
	ImportLocation    ImportLocation    `json:"import_location"`
}

func NewGameDataInfo(path string) IGameDataInfo {
	gamePart := NewInteraction().GamePart.GetGamePart()
	gameData, err := core.NewGameData(path, gamePart)
	if err != nil {
		l := logger.Get()
		l.Error().Err(err).Str("Path", path).Msg("Error creating game data")
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
	g.ExtractLocation.CreateTargetFileOutput(formatter, g)
	g.TranslateLocation.CreateTargetFileOutput(formatter, g)
	g.ImportLocation.GenerateTargetOutput(formatter, g)
}

func (g *GameDataInfo) GetGameDataInfo() *GameDataInfo {
	return g
}

func (g *GameDataInfo) GetGameData() *core.GameFiles {
	return &g.GameData
}

func (g *GameDataInfo) GetExtractLocation() *ExtractLocation {
	return &g.ExtractLocation
}

func (g *GameDataInfo) SetGameData(gameData *core.GameFiles) {
	g.GameData = *gameData
}

func (g *GameDataInfo) GetTranslateLocation() *TranslateLocation {
	return &g.TranslateLocation
}

func (g *GameDataInfo) GetImportLocation() *ImportLocation {
	return &g.ImportLocation
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
	targetPath := NewInteraction().GameLocation.GetTargetDirectory()

	if len(target) > 0 {
		targetPath = target[0]
	}

	if strings.HasPrefix(g.GameData.FullFilePath, targetPath) {
		g.GameData.RelativeGameDataPath = strings.TrimPrefix(g.GameData.FullFilePath, targetPath+string(os.PathSeparator))
	}
}
