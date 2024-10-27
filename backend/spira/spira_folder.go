package spira

import (
	"context"
	"ffxresources/backend/common"
	"ffxresources/backend/lib"
	"fmt"
)

type SpiraFolder struct {
	ctx      context.Context
	FileInfo *lib.FileInfo
}

func NewSpiraFolder(fileInfo *lib.FileInfo, extractPath, translatePath string) *SpiraFolder {
	/* extractedDirectory, err := lib.NewInteraction().ExtractLocation.ProvideTargetDirectory()
	if err != nil {
		lib.EmitError(ctx, err)
	}

	translatedDirectory, err := lib.GetWorkdirectory().ProvideTranslateLocation()
	if err != nil {
		lib.EmitError(ctx, err)
	} */

	/* relativePath, err := common.GetRelativePathFromMarker(fileInfo.AbsolutePath)
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	fileInfo.RelativePath = relativePath */

	fileInfo.ExtractLocation.TargetPath = common.PathJoin(extractPath, fileInfo.RelativePath)
	fileInfo.TranslateLocation.TargetPath = common.PathJoin(translatePath, fileInfo.RelativePath)

	return &SpiraFolder{
		ctx:      lib.NewInteraction().Ctx,
		FileInfo: fileInfo,
	}
}

func (d SpiraFolder) GetFileInfo() *lib.FileInfo {
	return d.FileInfo
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

func (d SpiraFolder) processFiles() []lib.IFileProcessor {
	var fileProcessors []lib.IFileProcessor

	results, err := common.EnumerateFilesDev(d.FileInfo.AbsolutePath)
	if err != nil {
		lib.NotifyError(err)
	}

	for _, result := range results {
		source, err := lib.NewSource(result)
		if err != nil {
			lib.NotifyError(err)
			continue
		}

		fileInfo := &lib.FileInfo{}
		lib.UpdateFileInfoFromSource(fileInfo, source)

		fileProcessor := NewFileProcessor(fileInfo)
		if fileProcessor == nil {
			lib.NotifyError(fmt.Errorf("invalid file type: %s", fileInfo.Name))
			continue
		}

		fileProcessors = append(fileProcessors, fileProcessor)
	}

	return fileProcessors
}
