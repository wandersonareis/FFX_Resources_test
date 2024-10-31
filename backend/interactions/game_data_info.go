package interactions

import "ffxresources/backend/core"

type GameDataInfo struct {
	GameData          core.GameData     `json:"game_data"`
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
		GameData:*gameData,
		ExtractLocation: *NewExtractLocation(),
		TranslateLocation: *NewTranslateLocation(),
		ImportLocation: *NewImportLocation(),
	}
}
