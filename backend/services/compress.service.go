package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"ffxresources/backend/models"
	"ffxresources/backend/spira"
	"fmt"
)

type CompressService struct{}

func NewCompressService() *CompressService {
	return &CompressService{}
}

func (c *CompressService) Compress(dataInfo *interactions.GameDataInfo) {
	if !common.FileExists(dataInfo.GameData.AbsolutePath) {
		lib.NotifyError(fmt.Errorf("original text file %s not found", dataInfo.GameData.Name))
		return
	}

	if !dataInfo.TranslateLocation.TargetFileExists() && dataInfo.GameData.Type != models.Dcp {
		lib.NotifyError(fmt.Errorf("translated text file %s not found", dataInfo.TranslateLocation.TargetFileName))
		return
	}

	err := core.EnsureWindowsLineBreaks(dataInfo.TranslateLocation.TargetFile, dataInfo.GameData.Type)
	if err != nil {
		lib.NotifyError(err)
		return
	}

	if core.CountSegments(dataInfo.TranslateLocation.TargetFile) < 0 {
		lib.NotifyError(fmt.Errorf("text file contains no separators: %s", dataInfo.GameData.Name))
		return
	}

	var fileProcessor models.ICompressor = spira.NewFileProcessor(dataInfo)
	if fileProcessor == nil {
		lib.NotifyError(fmt.Errorf("invalid file type: %s", dataInfo.GameData.Name))
		return
	}
	
	fileProcessor.Compress()
}
