package services

import (
	"ffxresources/backend/formats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
)

type ExtractService struct{}

func NewExtractService() *ExtractService {
	return &ExtractService{}
}

func (e *ExtractService) Extract(dataInfo *interactions.GameDataInfo) {
	fileProcessor := formatsDev.NewFileProcessor(dataInfo)
	if fileProcessor == nil {
		lib.NotifyError(fmt.Errorf("invalid file type: %s", dataInfo.GameData.Name))
		return
	}
	
	fileProcessor.Extract()
	lib.NotifySuccess("Extraction completed")
}
