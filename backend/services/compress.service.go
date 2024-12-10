package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	"ffxresources/backend/notifications"
	"fmt"
)

type CompressService struct{}

func NewCompressService() *CompressService {
	return &CompressService{}
}

func (c *CompressService) Compress(dataInfo *interactions.GameDataInfo) {
	if !common.IsFileExists(dataInfo.GameData.FullFilePath) {
		notifications.NotifyError(fmt.Errorf("game file %s not found", dataInfo.GameData.Name))
		return
	}

	gameData := dataInfo.GameData
	translateLocation := dataInfo.TranslateLocation

	if !gameData.IsDir {
		if err := translateLocation.Validate(); err != nil &&
			gameData.Type != models.Dcp {
			notifications.NotifyError(err)
			return
		}

		if err := common.EnsureWindowsLineBreaks(translateLocation.TargetFile, gameData.Type); err != nil {
			notifications.NotifyError(err)
			return
		}

		if common.CountSegments(translateLocation.TargetFile) < 0 {
			notifications.NotifyError(fmt.Errorf("text file %s is empty", gameData.Name))
			return
		}
	}

	fileProcessor := fileFormats.NewFileCompressor(dataInfo)
	if fileProcessor == nil {
		notifications.NotifyError(fmt.Errorf("invalid file type: %s", gameData.Name))
		return
	}

	if err := fileProcessor.Compress(); err != nil {
		notifications.NotifyError(err)
	}
}
