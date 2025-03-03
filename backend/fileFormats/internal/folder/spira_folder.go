package folder

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/baseFormats"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type SpiraFolder struct {
	baseFormats.IBaseFileFormat
	logger.ILoggerHandler
}

func NewSpiraFolder(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	return &SpiraFolder{
		IBaseFileFormat: baseFormats.NewFormatsBase(source, destination),

		ILoggerHandler: &logger.LogHandler{
			Logger: logger.Get().With().Str("module", "spira_folder").Logger(),
		},
	}
}

func (sf *SpiraFolder) Extract() error {
	return fmt.Errorf("use DirectoryExtractService instead")
}

func (sf *SpiraFolder) Compress() error {
	return fmt.Errorf("use DirectoryCompressService instead")
}

/* func (sf SpiraFolder) Extract() error {
	errChan := make(chan error)
	defer close(errChan)

	fileProcessors := sf.processFiles()

	progress := common.NewProgress(sf.Ctx)
	progress.SetMax(fileProcessors.GetLength())
	progress.Start()

	fileProcessors.ForEach(func(extractor interfaces.IFileProcessor) {
		if err := extractor.Extract(); err != nil {
			errChan <- err
		}

		progress.StepFile(extractor.GetSource().Get().Name)
	})

	progress.Stop()

	if err := <-errChan; err != nil {
		sf.LogError(err, "error extracting spira folder")
	}

	sf.LogInfo("Spira folder extracted", "folder", sf.GetSource().Get().Path)

	return nil
} */

/* func (sf SpiraFolder) Compress() error {
	errChan := make(chan error)
	defer close(errChan)

	fileProcessors := sf.processFiles()

	progress := common.NewProgress(sf.Ctx)
	progress.SetMax(fileProcessors.GetLength())
	progress.Start()

	fileProcessors.ForEach(func(compressor interfaces.IFileProcessor) {
		if err := compressor.Compress(); err != nil {
			errChan <- err
		}

		progress.Step()
	})

	progress.Stop()

	if err := <-errChan; err != nil {
		sf.LogError(err, "error compressing spira folder")
	}

	sf.LogInfo("Spira folder compressed", "folder", sf.GetSource().Get().Path)

	return nil
} */

/* func (sf SpiraFolder) processFiles() *components.List[interfaces.IFileProcessor] {
	filesList := components.NewEmptyList[string]()

	if err := components.ListFiles(filesList, sf.GetSource().Get().Path); err != nil {
		sf.LogError(err, "error listing files in directory", "directory", sf.GetSource().Get().Path)

		return components.NewEmptyList[interfaces.IFileProcessor]()
	}

	filesProcessorList := components.NewList[interfaces.IFileProcessor](filesList.GetLength())

	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()

	generateFilesProcessorListFunc := func(_ int, item string) {
		s, err := locations.NewSource(item, gameVersion)
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
} */
