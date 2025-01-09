package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/logger"
	"ffxresources/backend/notifications"
	"fmt"
)

type CompressService struct{}

func NewCompressService() *CompressService {
	return &CompressService{}
}

func (c *CompressService) Compress(file string) {
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

	if node, ok := nodeMap[file]; ok {
		fmt.Println(node)
		processor := node.Data.FileProcessor
		if processor != nil {
			if err := processor.Compress(); err != nil {
				notifications.NotifyError(err)
				return
			}
			notifications.NotifySuccess(fmt.Sprintf("File %s compressed successfully!", node.Label))
		}
	}

	/* if !common.IsFileExists(file) {
		notifications.NotifyError(fmt.Errorf("game file %s not found", common.GetFileName(file)))
		return
	}

	source, err := locations.NewSource(file, interactions.NewInteractionService().FFXGameVersion().GetGameVersion())
	if err != nil {
		notifications.NotifyError(err)
		return
	}

	destination := locations.NewDestination()
	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	translateLocation := destination.Translate().Get()

	sourceType := source.Get().Type
	if !source.Get().IsDir && sourceType != models.Dcp && sourceType != models.Lockit {
		if err := translateLocation.Validate(); err != nil {
			notifications.NotifyError(err)
			return
		}

		if err := common.EnsureWindowsLineBreaks(translateLocation.GetTargetFile()); err != nil {
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
		return
	}

	notifications.NotifySuccess(fmt.Sprintf("File %s compressed successfully!", source.Get().Name)) */
}
