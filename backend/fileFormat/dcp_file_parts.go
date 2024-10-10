package fileFormat

import (
	"context"
	"ffxresources/backend/lib"
)

type DcpFileParts struct {
	ctx      context.Context
	FileInfo lib.FileInfo
}

func NewDcpFileParts(ctx context.Context, fileInfo lib.FileInfo) *DcpFileParts {
	const (
		extractedFileExtension = ".txt"
		extractedDirectoryName = "system_text"
	)

	translatedDirectory, err := lib.GetWorkdirectory().ProvideTranslatedDirectory()
	if err != nil {
		lib.EmitError(ctx, err)
		return nil
	}

	extractedFile, extractedPath := generateDcpPartsExtractedOutput(fileInfo, extractedFileExtension)
	translatedFile, translatedPath := generateDcpTranslatedOutput(fileInfo, translatedDirectory, extractedDirectoryName)

	fileInfo.ExtractedFile = extractedFile
	fileInfo.ExtractedPath = extractedPath
	fileInfo.TranslatedFile = translatedFile
	fileInfo.TranslatedPath = translatedPath

	return &DcpFileParts{
		ctx:      ctx,
		FileInfo: fileInfo,
	}
}

func (d DcpFileParts) GetFileInfo() lib.FileInfo {
	return d.FileInfo
}

func (d DcpFileParts) Extract() {
	err := dialogsUnpacker(d.FileInfo)
	if err != nil {
		lib.EmitError(d.ctx, err)
		return
	}
}

func (d DcpFileParts) Compress() {
	err := dialogsTextPacker(d.FileInfo)
	if err != nil {
		lib.EmitError(d.ctx, err)
		return
	}
}

func generateDcpPartsExtractedOutput(fileInfo lib.FileInfo, targetExtension string) (string, string) {
	outputFile := lib.AddExtension(fileInfo.AbsolutePath, targetExtension)
	outputPath := fileInfo.Parent
	return outputFile, outputPath
}

func generateDcpTranslatedOutput(fileInfo lib.FileInfo, translatedDirectory, extractedDirectoryName string) (string, string) {
	outputFile := lib.PathJoin(translatedDirectory, extractedDirectoryName, fileInfo.Name)
	outputPath := lib.PathJoin(translatedDirectory, extractedDirectoryName)
	return outputFile, outputPath
}
