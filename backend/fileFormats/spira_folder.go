package fileFormats

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/interactions"
	"path/filepath"
)

type SpiraFolder struct {
	*base.FormatsBase
}

func NewSpiraFolder(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	gameFilesPath := interactions.NewInteraction().GameLocation.GetTargetDirectory()

	extractLocation := dataInfo.GetExtractLocation()
	translateLocation := dataInfo.GetTranslateLocation()

	relative := common.MakeRelativePath(dataInfo.GetGameData().FullFilePath, gameFilesPath)

	dataInfo.GetGameData().RelativeGameDataPath = relative

	dataInfo.GetExtractLocation().TargetPath = filepath.Join(extractLocation.TargetDirectory, relative)
	dataInfo.GetTranslateLocation().TargetPath = filepath.Join(translateLocation.TargetDirectory, relative)

	return &SpiraFolder{
		FormatsBase: base.NewFormatsBase(dataInfo),
	}
}

func (d SpiraFolder) Extract() {
	fileProcessors := d.processFiles()

	progress := common.NewProgress(d.Ctx)
	progress.SetMax(len(fileProcessors))
	progress.Start()

	worker := common.NewWorker[interactions.IFileProcessor]()

	worker.ParallelForEach(fileProcessors, func(_ int, fileProcessor interactions.IFileProcessor) {
		fileProcessor.Extract()

		progress.Step()
	})

	progress.Stop()
}

func (d SpiraFolder) Compress() {
	fileProcessors := d.processFiles()

	progress := common.NewProgress(d.Ctx)
	progress.SetMax(len(fileProcessors))
	progress.Start()

	worker := common.NewWorker[interactions.IFileProcessor]()

	worker.ParallelForEach(fileProcessors, func(_ int, fileProcessor interactions.IFileProcessor) {
		fileProcessor.Compress()

		progress.Step()
	})

	progress.Stop()
}

func (d SpiraFolder) processFiles() []interactions.IFileProcessor {
	results, err := common.ListFilesInDirectory(d.GetFileInfo().GetGameData().FullFilePath)
	if err != nil {
		d.Log.Error().Err(err).Msgf("error listing files in directory: %s", d.GetFileInfo().GetGameData().FullFilePath)
		return nil
	}

	var fileProcessors = make([]interactions.IFileProcessor, 0, len(results))

	worker := common.NewWorker[string]()

	worker.ParallelForEach(results, func(_ int, result string) {
		dataInfo := interactions.NewGameDataInfo(result)

		fileProcessor := NewFileProcessor(dataInfo)
		if fileProcessor == nil {
			d.Log.Error().Msgf("invalid file type: %s", dataInfo.GetGameData().Name)
			return
		}

		fileProcessors = append(fileProcessors, fileProcessor)
	})

	return fileProcessors
}
