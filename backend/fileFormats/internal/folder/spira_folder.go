package folder

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"ffxresources/backend/notifications"
	"path/filepath"

	"github.com/rs/zerolog"
)

type SpiraFolder struct {
	*base.FormatsBase
	fileProcessor func(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor

	log zerolog.Logger
}

func NewSpiraFolder(dataInfo interactions.IGameDataInfo, fileProcessor func(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor) interactions.IFileProcessor {
	gameFilesPath := interactions.NewInteraction().GameLocation.GetTargetDirectory()

	extractLocation := dataInfo.GetExtractLocation()
	translateLocation := dataInfo.GetTranslateLocation()

	relative := common.MakeRelativePath(dataInfo.GetGameData().FullFilePath, gameFilesPath)

	dataInfo.GetGameData().RelativeGameDataPath = relative

	dataInfo.GetExtractLocation().TargetPath = filepath.Join(extractLocation.TargetDirectory, relative)
	dataInfo.GetTranslateLocation().TargetPath = filepath.Join(translateLocation.TargetDirectory, relative)

	return &SpiraFolder{
		FormatsBase:   base.NewFormatsBase(dataInfo),
		fileProcessor: fileProcessor,

		log: logger.Get().With().Str("module", "spira_folder").Logger(),
	}
}

func (sf SpiraFolder) Extract() error {
	fileProcessors := sf.processFiles()

	progress := common.NewProgress(sf.Ctx)
	progress.SetMax(fileProcessors.GetLength())
	progress.Start()

	errChan := make(chan error, fileProcessors.GetLength())
	defer close(errChan)

	go notifications.ProcessError(errChan, sf.log)

	fileProcessors.ForEach(func(extractor interactions.IFileProcessor) {
		err := extractor.Extract()
		errChan <- err

		progress.Step()
	})

	progress.Stop()

	sf.Log.Info().
		Str("folder", sf.GetGameData().FullFilePath).
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

	fileProcessors.ForEach(func(compressor interactions.IFileProcessor) {
		err := compressor.Compress()
		errChan <- err

		progress.Step()
	})

	progress.Stop()

	sf.log.Info().
		Str("folder", sf.GetGameData().FullFilePath).
		Msg("Spira folder compressed")

	return nil
}

func (sf SpiraFolder) processFiles() *components.List[interactions.IFileProcessor] {
	//results, err := common.ListFilesInDirectoryDev(sf.GetFileInfo().GetGameData().FullFilePath)
	filesList := components.NewEmptyList[string]()
	err := components.ListFiles(filesList, sf.GetFileInfo().GetGameData().FullFilePath)
	if err != nil {
		sf.log.Error().
			Err(err).
			Str("directory", sf.GetFileInfo().GetGameData().FullFilePath).
			Msg("error listing files in directory")

		return components.NewEmptyList[interactions.IFileProcessor]()
	}

	filesProcessorList := components.NewList[interactions.IFileProcessor](filesList.GetLength())

	generateFilesProcessorListFunc := func(_ int, item string) {
		dataInfo := interactions.NewGameDataInfo(item)

		fileProcessor := sf.fileProcessor(dataInfo)
		if fileProcessor == nil {
			sf.Log.Error().
				Str("file", dataInfo.GetGameData().Name).
				Msg("invalid file type")

			return
		}

		filesProcessorList.Add(fileProcessor)
	}

	filesList.ParallelForEach(generateFilesProcessorListFunc)

	return filesProcessorList
}
