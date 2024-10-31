package services

import (
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"ffxresources/backend/models"
	"ffxresources/backend/spira"
	"fmt"
)

type ExtractService struct{}

func NewExtractService() *ExtractService {
	return &ExtractService{}
}

func (e *ExtractService) Extract(dataInfo *interactions.GameDataInfo) {
	var fileProcessor models.IExtractor = spira.NewFileProcessor(dataInfo)
	if fileProcessor == nil {
		lib.NotifyError(fmt.Errorf("invalid file type: %s", dataInfo.GameData.Name))
		return
	}
	
	fileProcessor.Extract()
	lib.NotifySuccess("Extraction completed")
}
