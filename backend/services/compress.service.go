package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/formats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"ffxresources/backend/models"
	"fmt"
)

type CompressService struct{}

func NewCompressService() *CompressService {
	return &CompressService{}
}

func (c *CompressService) Compress(dataInfo *interactions.GameDataInfo) {
	if !common.FileExists(dataInfo.GameData.AbsolutePath) {
		lib.NotifyError(fmt.Errorf("game file %s not found", dataInfo.GameData.Name))
		return
	}

	gameData := dataInfo.GameData
	translateLocation := dataInfo.TranslateLocation

	if !gameData.IsDir {
		if err := translateLocation.Validate(); err != nil &&
			gameData.Type != models.Dcp {
			lib.NotifyError(err)
			return
		}

		if err := core.EnsureWindowsLineBreaks(translateLocation.TargetFile, gameData.Type); err != nil {
			lib.NotifyError(err)
			return
		}

		if core.CountSegments(translateLocation.TargetFile) < 0 {
			lib.NotifyError(fmt.Errorf("text file %s is empty", gameData.Name))
			return
		}
	}

	fileProcessor := formatsDev.NewFileCompressor(dataInfo)
	if fileProcessor == nil {
		lib.NotifyError(fmt.Errorf("invalid file type: %s", gameData.Name))
		return
	}

	fileProcessor.Compress()
}
