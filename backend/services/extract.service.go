package services

import (
	"ffxresources/backend/events"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interactions"
	"fmt"
)

type ExtractService struct{}

func NewExtractService() *ExtractService {
	return &ExtractService{}
}

func (e *ExtractService) Extract(dataInfo *interactions.GameDataInfo) {
	fileProcessor := fileFormats.NewFileProcessor(dataInfo)
	if fileProcessor == nil {
		events.NotifyError(fmt.Errorf("invalid file type: %s", dataInfo.GameData.Name))
		return
	}
	
	fileProcessor.Extract()
	events.NotifySuccess("Extraction completed")
}
