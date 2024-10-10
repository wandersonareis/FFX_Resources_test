package spira

import (
	"context"
	"ffxresources/backend/lib"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type SpiraFolder struct {
	ctx      context.Context
	FileInfo lib.FileInfo
}

type Progress struct {
	Total      int `json:"total"`
	Processed  int `json:"processed"`
	Percentage int `json:"percentage"`
}

func NewSpiraFolder(ctx context.Context, fileInfo lib.FileInfo) SpiraFolder {
	extractedDirectory, err := lib.NewInteraction().ExtractLocation.ProvideTargetDirectory()
	if err != nil {
		lib.EmitError(ctx, err)
	}

	translatedDirectory, err := lib.GetWorkdirectory().ProvideTranslatedDirectory()
	if err != nil {
		lib.EmitError(ctx, err)
	}

	relativePath, err := lib.GetRelativePathFromMarker(fileInfo)
	if err != nil {
		lib.EmitError(ctx, err)
	}

	fileInfo.RelativePath = relativePath

	fileInfo.ExtractLocation = *lib.NewInteraction().ExtractLocation
	fileInfo.ExtractLocation.TargetPath = lib.PathJoin(extractedDirectory, relativePath)
	fileInfo.TranslatedPath = lib.PathJoin(translatedDirectory, relativePath)

	return SpiraFolder{
		ctx:      ctx,
		FileInfo: fileInfo,
	}
}

func (d SpiraFolder) GetFileInfo() lib.FileInfo {
	return d.FileInfo
}

func sendProgress(ctx context.Context, progress Progress) {
	runtime.EventsEmit(ctx, "Progress", progress)
}

func (d SpiraFolder) Extract() {
	fileProcessors := d.processFiles()
	totalFiles := len(fileProcessors)
	processedCount := 0

	sendProgress(d.ctx, Progress{
		Total:      totalFiles,
		Processed:  processedCount,
		Percentage: 0,
	})

	for _, fileProcessor := range fileProcessors {
		fileProcessor.Extract()

		processedCount++

		sendProgress(d.ctx, Progress{Total: totalFiles, Processed: processedCount, Percentage: processedCount * 100 / totalFiles})
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

	results, err := lib.EnumerateFilesDev(d.FileInfo.AbsolutePath)
	if err != nil {
		fmt.Println("Error:", err)
		lib.EmitError(d.ctx, err)
	}

	for _, result := range results {
		source, err := lib.NewSource(result)
		if err != nil {
			fmt.Println("Error:", err)
			lib.EmitError(d.ctx, err)
			continue
		}

		fileInfo, err := lib.CreateFileInfo(source)
		if err != nil {
			lib.EmitError(d.ctx, err)
			continue
		}

		fileProcessor := NewFileProcessor(d.ctx, fileInfo)
		if fileProcessor == nil {
			lib.EmitError(d.ctx, fmt.Errorf("invalid file type: %s", fileInfo.Name))
			continue
		}

		fileProcessors = append(fileProcessors, fileProcessor)
	}

	return fileProcessors
}
