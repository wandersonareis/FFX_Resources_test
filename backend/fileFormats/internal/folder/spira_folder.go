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
	"path/filepath"

	"github.com/rs/zerolog"
)

type SpiraFolder struct {
	*base.FormatsBase
	fileProcessor func(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor

	log zerolog.Logger
}

func NewSpiraFolder(source interfaces.ISource, destination locations.IDestination, fileProcessor func(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor) interfaces.IFileProcessor {
	gameFilesPath := interactions.NewInteraction().GameLocation.GetTargetDirectory()

	extractLocation := destination.Extract().Get()
	translateLocation := destination.Translate().Get()

	relative := common.MakeRelativePath(source.Get().Path, gameFilesPath)

	source.Get().RelativePath = relative

	destination.Extract().Get().SetTargetPath(filepath.Join(extractLocation.GetTargetDirectory(), relative))
	destination.Translate().Get().SetTargetPath(filepath.Join(translateLocation.GetTargetDirectory(), relative))

	return &SpiraFolder{
		FormatsBase:   base.NewFormatsBaseDev(source, destination),
		fileProcessor: fileProcessor,

		log: logger.Get().With().Str("module", "spira_folder").Logger(),
	}
}

func (sf SpiraFolder) Extract() error {
	errChan := make(chan error)
	go notifications.ProcessError(errChan, sf.log)
	go notifications.ProcessError(errChan, sf.log)
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

	sf.Log.Info().
		Str("folder", sf.Source().Get().Path).
		Msg("Spira folder extracted")

	return nil
}

func (sf SpiraFolder) Compress() error {
	fileProcessors := sf.processFiles()

	progress := common.NewProgress(sf.Ctx)
	progress.SetMax(fileProcessors.GetLength())
	progress.Start()

	errChan := make(chan error, fileProcessors.GetLength())
	defer close(errChan)

	go notifications.ProcessError(errChan, sf.log)

	fileProcessors.ForEach(func(compressor interfaces.IFileProcessor) {
		err := compressor.Compress()
		errChan <- err

		progress.Step()
	})

	progress.Stop()

	sf.log.Info().
		Str("folder", sf.Source().Get().Path).
		Msg("Spira folder compressed")

	return nil
}

func (sf SpiraFolder) processFiles() *components.List[interfaces.IFileProcessor] {
	filesList := components.NewEmptyList[string]()
	err := components.ListFiles(filesList, sf.Source().Get().Path)
	if err != nil {
		sf.log.Error().
			Err(err).
			Str("directory", sf.Source().Get().Path).
			Msg("error listing files in directory")

		return components.NewEmptyList[interfaces.IFileProcessor]()
	}

	filesProcessorList := components.NewList[interfaces.IFileProcessor](filesList.GetLength())

	generateFilesProcessorListFunc := func(_ int, item string) {
		s, err := locations.NewSource(item, interactions.Get().GamePart.GetGamePart())
		if err != nil {
			sf.log.Error().
				Err(err).
				Str("file", item).
				Msg("error creating source")
		}

		t := locations.NewDestination()

		fileProcessor := sf.fileProcessor(s, t)
		if fileProcessor == nil {
			sf.Log.Error().
				Str("file", s.Get().Name).
				Msg("invalid file type")

			return
		}

		filesProcessorList.Add(fileProcessor)
	}

	filesList.ParallelForEach(generateFilesProcessorListFunc)

	return filesProcessorList
}
