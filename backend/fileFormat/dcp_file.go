package fileFormat

import (
	"context"
	"ffxresources/backend/lib"
	"fmt"
	"path/filepath"
	"time"
)

const xplitedFileName = "macrodic"
const pattern = "macrodic\\..*?\\.00[0-6]"

type DcpFile struct {
	ctx      context.Context
	FileInfo lib.FileInfo
}

func NewDcpFile(ctx context.Context, fileInfo lib.FileInfo) *DcpFile {
	relativePath, err := lib.GetRelativePathFromMarker(fileInfo)
	if err != nil {
		lib.EmitError(ctx, err)
		return nil
	}

	fileInfo.RelativePath = relativePath

	translatedDirectory, err := lib.GetWorkdirectory().ProvideTranslatedDirectory()
	if err != nil {
		lib.EmitError(ctx, err)
		return nil
	}
	
	translatedFile, translatedPath := lib.GeneratedTranslatedOutput(fileInfo, translatedDirectory)

	fileInfo.TranslatedFile = translatedFile
	fileInfo.TranslatedPath = translatedPath

	fileInfo.ExtractLocation = *lib.NewInteraction().ExtractLocation
	fileInfo.ExtractLocation.GenerateTargetOutput(TxtFormatter{}, fileInfo)

	return &DcpFile{
		ctx:      ctx,
		FileInfo: fileInfo,
	}
}

func (d DcpFile) GetFileInfo() lib.FileInfo {
	return d.FileInfo
}

func (d DcpFile) Extract() {
	err := dcpFileXpliter(d.GetFileInfo())
	if err != nil {
		lib.EmitError(d.ctx, err)
		return
	}

	time.Sleep(500 * time.Millisecond)
	macrodicPath := d.FileInfo.ExtractLocation.TargetPath
	xplitedFiles, err := lib.EnumerateFilesByPattern(macrodicPath, pattern)
	if err != nil {
		lib.EmitError(d.ctx, err)
		return
	}

	if len(xplitedFiles) < 7 {
		lib.EmitError(d.ctx, fmt.Errorf("invalid number of xplited files: %d", len(xplitedFiles)))
		return
	}

	for _, xplitedFile := range xplitedFiles {
		source, err := lib.NewSource(xplitedFile)
		if err != nil {
			lib.EmitError(d.ctx, err)
			return
		}

		err = lib.EnsurePathExists(source.Parent)
		if err != nil {
			lib.EmitError(d.ctx, err)
			return
		}

		fileInfo, err := lib.CreateFileInfo(source)
		if err != nil {
			lib.EmitError(d.ctx, err)
			return
		}

		fileInfo.ExtractLocation = *lib.NewInteraction().ExtractLocation
		fileInfo.ExtractLocation.TargetFile = lib.AddExtension(fileInfo.AbsolutePath, lib.DEFAULT_TEXT_EXTENSION)
		fileInfo.ExtractLocation.TargetPath = filepath.Dir(fileInfo.ExtractLocation.TargetFile)

		/* fileInfo.ExtractedFile = lib.AddExtension(fileInfo.AbsolutePath, lib.DEFAULT_TEXT_EXTENSION)
		fileInfo.ExtractedPath = filepath.Dir(fileInfo.ExtractedFile) */

		if fileInfo.ExtractLocation.TargetFileExists() {
			continue
		}

		textFile := NewDcpFileParts(d.ctx, fileInfo)
		if textFile == nil {
			lib.InvalidFileType(d.ctx, fileInfo.Name)
			return
		}

		textFile.Extract()
	}
}

func (d DcpFile) Compress() {
	var translatedPath string

	macrodicPath := d.FileInfo.ExtractedPath

	macrodicFilesPattern := pattern + "$"
	macrodicTextFilesPattern := pattern + "\\.txt"

	xplitedFiles, err := lib.EnumerateFilesByPattern(macrodicPath, macrodicFilesPattern)
	if err != nil {
		lib.EmitError(d.ctx, err)
		return
	}

	if len(xplitedFiles) < 7 {
		lib.EmitError(d.ctx, fmt.Errorf("invalid number of xplited files: %d", len(xplitedFiles)))
		return
	}

	xplitedTextFiles, err := lib.EnumerateFilesByPattern(macrodicPath, macrodicTextFilesPattern)
	if err != nil {
		lib.EmitError(d.ctx, err)
		return
	}

	if len(xplitedTextFiles) < 7 {
		lib.EmitError(d.ctx, fmt.Errorf("invalid number of xplited text files: %d", len(xplitedTextFiles)))
		return
	}

	for i := 0; i < len(xplitedFiles); i++ {
		macrodicPart := xplitedFiles[i]

		sourcePart, sourcePartErr := lib.NewSource(macrodicPart)
		if sourcePartErr != nil {
			lib.EmitError(d.ctx, err)
			return
		}

		partFileInfo, err := lib.CreateFileInfo(sourcePart)
		if err != nil {
			lib.EmitError(d.ctx, err)
			return
		}

		fileHandle := NewDcpFileParts(d.ctx, partFileInfo)
		if fileHandle == nil {
			lib.InvalidFileType(d.ctx, partFileInfo.Name)
			return
		}
		fileHandle.Compress()

		if translatedPath == "" {
			translatedPath = lib.PathJoin(fileHandle.FileInfo.TranslatedPath, xplitedFileName)
		}
	}

	err = dcpFileJoiner(d.FileInfo, translatedPath)
	if err != nil {
		lib.EmitError(d.ctx, err)
		return
	}
}

/* func generateDcpExtractedOutput(workDirectory, targetDirName, targetFileName string) (string, string) {
	outputFile := filepath.Join(workDirectory, targetDirName, targetFileName)
	outputPath := filepath.Dir(outputFile)
	return outputFile, outputPath
} */
