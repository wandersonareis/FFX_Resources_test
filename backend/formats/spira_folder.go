package formatsDev

import (
	"context"
	"ffxresources/backend/common"
	"ffxresources/backend/events"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
	"path/filepath"
)

type SpiraFolder struct {
	ctx      context.Context
	DataInfo *interactions.GameDataInfo
}

func NewSpiraFolder(dataInfo *interactions.GameDataInfo, extractPath, translatePath string) *SpiraFolder {
	dataInfo.ExtractLocation.TargetPath = filepath.Join(extractPath, dataInfo.GameData.RelativeGameDataPath)
	dataInfo.TranslateLocation.TargetPath = filepath.Join(translatePath, dataInfo.GameData.RelativeGameDataPath)

	return &SpiraFolder{
		ctx:      interactions.NewInteraction().Ctx,
		DataInfo: dataInfo,
	}
}

func (d SpiraFolder) GetFileInfo() *interactions.GameDataInfo {
	return d.DataInfo
}

func (d SpiraFolder) Extract() {
	fileProcessors := d.processFiles()
	totalFiles := len(fileProcessors)
	processedCount := 0

	lib.SendProgress(d.ctx, lib.Progress{
		Total:      totalFiles,
		Processed:  processedCount,
		Percentage: 0,
	})

	lib.ShowProgressBar(d.ctx)

	worker := common.NewWorker[interactions.IFileProcessor]()

	worker.ParallelForEach(fileProcessors, func(_ int, fileProcessor interactions.IFileProcessor) {
		fileProcessor.Extract()
	})
}

func (d SpiraFolder) Compress() {
	fileProcessors := d.processFiles()

	worker := common.NewWorker[interactions.IFileProcessor]()

	worker.ParallelForEach(fileProcessors, func(_ int, fileProcessor interactions.IFileProcessor) {
		fileProcessor.Compress()
	})
}

func (d SpiraFolder) processFiles() []interactions.IFileProcessor {
	results, err := common.ListFilesInDirectory(d.DataInfo.GameData.FullFilePath)
	if err != nil {
		events.NotifyError(err)
		return nil
	}

	var fileProcessors = make([]interactions.IFileProcessor, 0, len(results))

	worker := common.NewWorker[string]()

	worker.ParallelForEach(results, func(_ int, result string) {
		dataInfo := interactions.NewGameDataInfo(result)

		fileProcessor := NewFileProcessor(dataInfo)
		if fileProcessor == nil {
			events.NotifyError(fmt.Errorf("invalid file type: %s", dataInfo.GameData.Name))
			return
		}

		fileProcessors = append(fileProcessors, fileProcessor)
	})

	return fileProcessors
}
