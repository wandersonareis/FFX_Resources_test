package services

import (
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"
)

type ExtractService struct{}

func NewExtractService() *ExtractService {
	return &ExtractService{}
}

func (e *ExtractService) Extract(dataInfo *interactions.GameDataInfo) {
	fileProcessor := fileFormats.NewFileProcessor(dataInfo)
	if fileProcessor == nil {
		l := logger.Get()
		l.Error().Err(fmt.Errorf("invalid file type: %s", dataInfo.GameData.Name)).Msg("Error extracting file")
		return
	}
	
	fileProcessor.Extract()
}
