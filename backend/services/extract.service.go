package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"ffxresources/backend/notifications"
	"fmt"
)

type ExtractService struct{}

func NewExtractService() *ExtractService {
	return &ExtractService{}
}

func (e *ExtractService) Extract(file string) {
	defer func() {
		if r := recover(); r != nil {
			l := logger.Get()
			l.Error().
				Interface("recover", r).
				Str("file", common.GetFileName(file)).
				Msg("Panic occurred during extraction")

			notifications.NotifyError(fmt.Errorf("panic occurred: %v", r))
		}
	}()

	source, err := locations.NewSource(file, interactions.NewInteraction().FFXGameVersion().GetGameVersion())
	if err != nil {
		notifications.NotifyError(err)
		return
	}

	destination := locations.NewDestination()
	destination.InitializeLocations(source, formatters.NewTxtFormatterDev())

	fileProcessor := fileFormats.NewFileProcessor(source, destination)
	if fileProcessor == nil {
		l := logger.Get()
		l.Error().
			Err(fmt.Errorf("invalid file type: %s", source.Get().Name)).
			Msg("Error extracting file")

		return
	}

	if err := fileProcessor.Extract(); err != nil {
		notifications.NotifyError(err)
	}
}
