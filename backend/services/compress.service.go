package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"ffxresources/backend/models"
	"ffxresources/backend/notifications"
	"fmt"
)

type CompressService struct{}

func NewCompressService() *CompressService {
	return &CompressService{}
}

func (c *CompressService) Compress(source interfaces.ISource, destination locations.IDestination) {
	defer func() {
		if r := recover(); r != nil {
			l := logger.Get()
			l.Error().
				Interface("recover", r).
				Str("file", source.Get().Name).
				Msg("Panic occurred during extraction")

			notifications.NotifyError(fmt.Errorf("panic occurred: %v", r))
		}
	}()

	if !common.IsFileExists(source.Get().Path) {
		notifications.NotifyError(fmt.Errorf("game file %s not found", source.Get().Name))
		return
	}

	translateLocation := destination.Translate().Get()

	if !source.Get().IsDir {
		if err := translateLocation.Validate(); err != nil &&
			source.Get().Type != models.Dcp {
			notifications.NotifyError(err)
			return
		}

		if err := common.EnsureWindowsLineBreaks(translateLocation.GetTargetFile(), source.Get().Type); err != nil {
			notifications.NotifyError(err)
			return
		}

		if common.CountSegments(translateLocation.GetTargetFile()) < 0 {
			notifications.NotifyError(fmt.Errorf("text file %s is empty", source.Get().Name))
			return
		}
	}

	fileProcessor := fileFormats.NewFileCompressor(source, destination)
	if fileProcessor == nil {
		notifications.NotifyError(fmt.Errorf("invalid file type: %s", source.Get().Name))
		return
	}

	if err := fileProcessor.Compress(); err != nil {
		notifications.NotifyError(err)
	}
}
