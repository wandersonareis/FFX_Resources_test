package services

import (
	"ffxresources/backend/common"
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
		l := logger.Get()
		if r := recover(); r != nil {
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
			if err := processor.Extract(); err != nil {
				notifications.NotifyError(err)
				return
			}
			notifications.NotifySuccess(fmt.Sprintf("File %s extracted successfully!", node.Label))
		}
	}

	/* source, err := locations.NewSource(file, interactions.NewInteractionService().FFXGameVersion().GetGameVersion())
	if err != nil {
		notifications.NotifyError(err)
		return
	}

	destination := locations.NewDestination()
	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	fileProcessor := fileFormats.NewFileExtractor(source, destination)
	if fileProcessor == nil {
		l := logger.Get()
		l.Error().
			Err(fmt.Errorf("invalid file type: %s", source.Get().Name)).
			Msg("Error extracting file")

		return
	}

	if err := fileProcessor.Extract(); err != nil {
		notifications.NotifyError(err)
		return
	} */

	//notifications.NotifySuccess(fmt.Sprintf("File %s extracted successfully!", source.Get().Name))
}
