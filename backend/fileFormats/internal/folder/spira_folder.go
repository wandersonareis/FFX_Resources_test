package folder

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
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

func (sf SpiraFolder) Extract() {
	fileProcessors := sf.processFiles()

	progress := common.NewProgress(sf.Ctx)
	progress.SetMax(len(fileProcessors))
	progress.Start()

	worker := common.NewWorker[interactions.IFileProcessor]()

	worker.ParallelForEach(&fileProcessors, func(_ int, fileProcessor interactions.IFileProcessor) {
		fileProcessor.Extract()

		progress.Step()
	})

	progress.Stop()

	sf.Log.Info().
		Str("folder", sf.GetGameData().FullFilePath).
		Msg("Spira folder extracted")
}

func (sf SpiraFolder) Compress() {
	fileProcessors := sf.processFiles()

	progress := common.NewProgress(sf.Ctx)
	progress.SetMax(len(fileProcessors))
	progress.Start()

	worker := common.NewWorker[interactions.IFileProcessor]()

	worker.ParallelForEach(&fileProcessors, func(_ int, fileProcessor interactions.IFileProcessor) {
		fileProcessor.Compress()

		progress.Step()
	})

	progress.Stop()

	sf.log.Info().
		Str("folder", sf.GetGameData().FullFilePath).
		Msg("Spira folder compressed")
}

func (sf SpiraFolder) processFiles() []interactions.IFileProcessor {
	results, err := common.ListFilesInDirectory(sf.GetFileInfo().GetGameData().FullFilePath)
	if err != nil {
		sf.log.Error().
			Err(err).
			Str("directory", sf.GetFileInfo().GetGameData().FullFilePath).
			Msg("error listing files in directory")

		return nil
	}

	var fileProcessors = make([]interactions.IFileProcessor, 0, len(*results))

	worker := common.NewWorker[string]()

	worker.ParallelForEach(results, func(_ int, result string) {
		dataInfo := interactions.NewGameDataInfo(result)

		fileProcessor := sf.fileProcessor(dataInfo)
		if fileProcessor == nil {
			sf.Log.Error().
				Str("file", dataInfo.GetGameData().Name).
				Msg("invalid file type")

			return
		}

		fileProcessors = append(fileProcessors, fileProcessor)
	})

	return fileProcessors
}
