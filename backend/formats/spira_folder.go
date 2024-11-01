package formats

import (
	"context"
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
)

type SpiraFolder struct {
	ctx      context.Context
	DataInfo *interactions.GameDataInfo
}

func NewSpiraFolder(dataInfo *interactions.GameDataInfo, extractPath, translatePath string) *SpiraFolder {
	dataInfo.ExtractLocation.TargetPath = common.PathJoin(extractPath, dataInfo.GameData.RelativePath)
	dataInfo.TranslateLocation.TargetPath = common.PathJoin(translatePath, dataInfo.GameData.RelativePath)

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

	for _, fileProcessor := range fileProcessors {
		fileProcessor.Extract()

		processedCount++

		lib.SendProgress(d.ctx, lib.Progress{Total: totalFiles, Processed: processedCount, Percentage: processedCount * 100 / totalFiles})
	}
}

func (d SpiraFolder) Compress() {
	fileProcessors := d.processFiles()

	for _, fileProcessor := range fileProcessors {
		fileProcessor.Compress()
	}
}

func (d SpiraFolder) processFiles() []interactions.IFileProcessor {
	results, err := common.EnumerateFilesDev(d.DataInfo.GameData.AbsolutePath)
	if err != nil {
		lib.NotifyError(err)
	}

	var fileProcessors = make([]interactions.IFileProcessor, 0, len(results))

	for _, result := range results {
		/* source, err := lib.NewSource(result)
		if err != nil {
			lib.NotifyError(err)
			continue
		} */

		/* fileInfo := &lib.FileInfo{}
		lib.UpdateFileInfoFromSource(fileInfo, source)
		*/
		dataInfo := interactions.NewGameDataInfo(result)

		fileProcessor := NewFileProcessor(dataInfo)
		if fileProcessor == nil {
			lib.NotifyError(fmt.Errorf("invalid file type: %s", dataInfo.GameData.Name))
			continue
		}

		fileProcessors = append(fileProcessors, fileProcessor)
	}

	return fileProcessors
}
