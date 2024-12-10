package services

import (
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"ffxresources/backend/notifications"
	"fmt"
)

type ExtractService struct{}

func NewExtractService() *ExtractService {
	return &ExtractService{}
}

func (e *ExtractService) Extract(dataInfo *interactions.GameDataInfo) {
	defer func() {
		if r := recover(); r != nil {
			l := logger.Get()
			l.Error().
				Interface("recover", r).
				Str("file", dataInfo.GameData.Name).
				Msg("Panic occurred during extraction")

			notifications.NotifyError(fmt.Errorf("panic occurred: %v", r))
		}
	}()

	fileProcessor := fileFormats.NewFileProcessor(dataInfo)
	if fileProcessor == nil {
		l := logger.Get()
		l.Error().
			Err(fmt.Errorf("invalid file type: %s", dataInfo.GameData.Name)).
			Msg("Error extracting file")
			
		return
	}

	if err := fileProcessor.Extract(); err != nil {
		notifications.NotifyError(err)
	}
}
