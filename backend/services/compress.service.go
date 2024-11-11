package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/events"
	formatsDev "ffxresources/backend/formats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	"fmt"
)

type CompressService struct{}

func NewCompressService() *CompressService {
	return &CompressService{}
}

func (c *CompressService) Compress(dataInfo *interactions.GameDataInfo) {
	if !common.IsFileExists(dataInfo.GameData.FullFilePath) {
		events.NotifyError(fmt.Errorf("game file %s not found", dataInfo.GameData.Name))
		return
	}

	gameData := dataInfo.GameData
	translateLocation := dataInfo.TranslateLocation

	if !gameData.IsDir {
		if err := translateLocation.Validate(); err != nil &&
			gameData.Type != models.Dcp {
			events.NotifyError(err)
			return
		}

		if err := core.EnsureWindowsLineBreaks(translateLocation.TargetFile, gameData.Type); err != nil {
			events.NotifyError(err)
			return
		}

		if core.CountSegments(translateLocation.TargetFile) < 0 {
			events.NotifyError(fmt.Errorf("text file %s is empty", gameData.Name))
			return
		}
	}

	fileProcessor := formatsDev.NewFileCompressor(dataInfo)
	if fileProcessor == nil {
		events.NotifyError(fmt.Errorf("invalid file type: %s", gameData.Name))
		return
	}

	fileProcessor.Compress()
}
