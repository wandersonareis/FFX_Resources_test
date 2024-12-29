package folder

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"ffxresources/backend/notifications"
)

type SpiraFolder struct {
	*base.FormatsBase
	logger.ILoggerHandler

	fileProcessor func(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor
}

func NewSpiraFolder(source interfaces.ISource, destination locations.IDestination, fileProcessor func(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor) interfaces.IFileProcessor {
	return &SpiraFolder{
		FormatsBase:   base.NewFormatsBase(source, destination),
		ILoggerHandler: &logger.LogHandler{
			Logger: logger.Get().With().Str("module", "spira_folder").Logger(),
		},
		fileProcessor: fileProcessor,
	}
}

func (sf SpiraFolder) Extract() error {
	errChan := make(chan error)
	go notifications.ProcessError(errChan, sf.GetLogger())
	go notifications.ProcessError(errChan, sf.GetLogger())
	defer close(errChan)

	fileProcessors := sf.processFiles()

	progress := common.NewProgress(sf.Ctx)
	progress.SetMax(fileProcessors.GetLength())
	progress.Start()

	fileProcessors.ForEach(func(extractor interfaces.IFileProcessor) {
		err := extractor.Extract()
		errChan <- err

		progress.StepFile(extractor.Source().Get().Name)
	})

	progress.Stop()

	sf.LogInfo("Spira folder extracted", "folder", sf.Source().Get().Path)

	return nil
}

func (sf SpiraFolder) Compress() error {
	errChan := make(chan error)
	go notifications.ProcessError(errChan, sf.GetLogger())
	defer close(errChan)

	fileProcessors := sf.processFiles()

	progress := common.NewProgress(sf.Ctx)
	progress.SetMax(fileProcessors.GetLength())
	progress.Start()

	fileProcessors.ForEach(func(compressor interfaces.IFileProcessor) {
		err := compressor.Compress()
		errChan <- err

		progress.Step()
	})

	progress.Stop()

	sf.LogInfo("Spira folder compressed", "folder", sf.Source().Get().Path)

	return nil
}

func (sf SpiraFolder) processFiles() *components.List[interfaces.IFileProcessor] {
	filesList := components.NewEmptyList[string]()
	err := components.ListFiles(filesList, sf.Source().Get().Path)
	if err != nil {
		sf.LogError(err, "error listing files in directory", "directory", sf.Source().Get().Path)

		return components.NewEmptyList[interfaces.IFileProcessor]()
	}

	filesProcessorList := components.NewList[interfaces.IFileProcessor](filesList.GetLength())

	generateFilesProcessorListFunc := func(_ int, item string) {
		s, err := locations.NewSource(item, interactions.NewInteractionService().FFXGameVersion().GetGameVersion())
		if err != nil {
			sf.LogError(err, "error creating source", "file", item)
		}

		t := locations.NewDestination()

		fileProcessor := sf.fileProcessor(s, t)
		if fileProcessor == nil {
			sf.LogError(nil, "invalid file type", "file", s.Get().Name)

			return
		}

		filesProcessorList.Add(fileProcessor)
	}

	filesList.ParallelForEach(generateFilesProcessorListFunc)

	return filesProcessorList
}
