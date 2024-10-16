package formats

import (
	"ffxresources/backend/lib"
	"path/filepath"
)

type TxtFormatter struct{}

func (t TxtFormatter) ReadFile(fileInfo *lib.FileInfo, targetDirectory string) (string, string) {
	var outputFile, outputPath string

	switch fileInfo.Type {
	case lib.Dcp:
		outputFile, outputPath = t.provideDcpReadPath(fileInfo, targetDirectory)
	case lib.DcpParts:
		outputFile, outputPath = t.provideDcpPartsReadPath(fileInfo, targetDirectory)
	default:
		outputFile, outputPath = t.provideDefaulReadPath(fileInfo, targetDirectory)
	}

	return outputFile, outputPath
}

func (t TxtFormatter) provideDefaulReadPath(fileInfo *lib.FileInfo, targetDirectory string) (string, string) {
	//outputFile := PathJoin(workDirectory, targetDirName, ChangeExtension(fileInfo.RelativePath, targetExtension))

	/* extractedFile, extractedPath := lib.GenerateExtractedOutput(fileInfo, targetDirectory, "", lib.DEFAULT_TEXT_EXTENSION)

	return extractedFile, extractedPath */

	return provideBasePath(targetDirectory, lib.ChangeExtension(fileInfo.RelativePath, lib.DEFAULT_TEXT_EXTENSION))
}

func (t TxtFormatter) provideDcpReadPath(fileInfo *lib.FileInfo, targetDirectory string) (string, string) {
	outputFile := lib.PathJoin(targetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME, fileInfo.Name)

	outputPath := lib.PathJoin(targetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME)

	return outputFile, outputPath
}

func (t TxtFormatter) provideDcpPartsReadPath(fileInfo *lib.FileInfo, targetDirectory string) (string, string) {
	return provideBasePath(targetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME, lib.AddExtension(fileInfo.Name, lib.DEFAULT_TEXT_EXTENSION))
}

func (t TxtFormatter) WriteFile(fileInfo *lib.FileInfo, targetDirectory string) (string, string) {

	var outputFile, outputPath string

	switch fileInfo.Type {
	case lib.Dcp:
		outputFile, outputPath = t.provideDcpWritePath(fileInfo, targetDirectory)
	case lib.DcpParts:
		outputFile, outputPath = t.provideDcpPartsWritePath(fileInfo, targetDirectory)
	default:
		outputFile, outputPath = t.provideDefaultWritePath(fileInfo, targetDirectory)
	}

	return outputFile, outputPath
}

func (t TxtFormatter) provideDefaultWritePath(fileInfo *lib.FileInfo, targetDirectory string) (string, string) {
	/* outputFile := lib.PathJoin(targetDirectory, lib.ChangeExtension(fileInfo.RelativePath, fileInfo.Extension))
	outputPath := filepath.Dir(outputFile)

	return outputFile, outputPath */
	return provideBasePath(targetDirectory, lib.ChangeExtension(fileInfo.RelativePath, fileInfo.Extension))
}

func (t TxtFormatter) provideDcpWritePath(fileInfo *lib.FileInfo, targetDirectory string) (string, string) {
	/* outputFile := lib.PathJoin(targetDirectory, fileInfo.RelativePath)
	outputPath := lib.GetDir(outputFile)

	return outputFile, outputPath */
	return provideBasePath(targetDirectory, fileInfo.RelativePath)
}

func (t TxtFormatter) provideDcpPartsWritePath(fileInfo *lib.FileInfo, targetDirectory string) (string, string) {
	/* outputFile := lib.PathJoin(targetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME, fileInfo.Name)

	outputPath := lib.GetDir(outputFile)

	return outputFile, outputPath */
	return provideBasePath(targetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME, fileInfo.Name)
}

func provideBasePath(targetDirectory string, dirParts ...string) (string, string) {
	outputFile := lib.PathJoin(targetDirectory, filepath.Join(dirParts...))
	outputPath := filepath.Dir(outputFile)

	return outputFile, outputPath
}