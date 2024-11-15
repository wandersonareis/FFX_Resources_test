package util

import "ffxresources/backend/interactions"

func ErrorObject(info interactions.IGameDataInfo) map[string]interface{} {
	return map[string]interface{}{
		"gameData": info.GetGameData(),
		"extract":  info.GetExtractLocation(),
		"translate": info.GetTranslateLocation(),
	}
}