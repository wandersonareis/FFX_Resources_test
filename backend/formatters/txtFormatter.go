package formatters

import (
	"ffxresources/backend/common"
	"ffxresources/backend/formats/lib"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	"path/filepath"
)

type TxtFormatter struct {
	targetExtension string
}

func NewTxtFormatter() *TxtFormatter {
	return &TxtFormatter{
		targetExtension: ".txt",
	}
}

func (t TxtFormatter) ReadFile(dataInfo *interactions.GameDataInfo, targetDirectory string) (string, string) {
	var outputFile, outputPath string

	switch dataInfo.GameData.Type {
	case models.Dcp:
		outputFile, outputPath = t.provideDcpReadPath(targetDirectory, dataInfo.GameData.Name)
	case models.DcpParts:
		outputFile, outputPath = t.providePartsReadPath(targetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME, dataInfo.GameData.Name)
	case models.Lockit:
		outputFile, outputPath = t.provideLockitReadPath(targetDirectory, dataInfo.GameData.NamePrefix)
	case models.LockitParts:
		outputFile, outputPath = t.providePartsReadPath(targetDirectory, lib.LOCKIT_TARGET_DIR_NAME, dataInfo.GameData.Name)
	default:
		outputFile, outputPath = t.provideDefaulReadPath(targetDirectory, dataInfo.GameData.RelativeGameDataPath)
	}

	return outputFile, outputPath
}

func (t TxtFormatter) provideDefaulReadPath(targetDirectory, relativePath string) (string, string) {
	return provideBasePath(targetDirectory, common.ChangeExtension(relativePath, t.targetExtension))
}

func (t TxtFormatter) provideDcpReadPath(targetDirectory, fileName string) (string, string) {
	outputFile := filepath.Join(targetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME, fileName)

	outputPath := filepath.Join(targetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME)

	return outputFile, outputPath
}

func (t TxtFormatter) providePartsReadPath(targetDirectory, dirName, fileName string) (string, string) {
	return provideBasePath(targetDirectory, dirName, common.AddExtension(fileName, t.targetExtension))
}

func (t TxtFormatter) provideLockitReadPath(targetDirectory, fileName string) (string, string) {
	return provideBasePath(targetDirectory, lib.LOCKIT_TARGET_DIR_NAME, common.AddExtension(fileName, t.targetExtension))
}

func (t TxtFormatter) WriteFile(fileInfo *interactions.GameDataInfo, targetDirectory string) (string, string) {

	var outputFile, outputPath string

	switch fileInfo.GameData.Type {
	case models.Dcp:
		outputFile, outputPath = t.provideDcpWritePath(targetDirectory, fileInfo.GameData.RelativeGameDataPath)
	case models.DcpParts:
		outputFile, outputPath = t.providePartsWritePath(targetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME, fileInfo.GameData.Name)
	case models.LockitParts:
		outputFile, outputPath = t.providePartsWritePath(targetDirectory, lib.LOCKIT_TARGET_DIR_NAME, fileInfo.GameData.Name)
	default:
		outputFile, outputPath = t.provideDefaultWritePath(targetDirectory, fileInfo.GameData.RelativeGameDataPath, fileInfo.GameData.Extension)
	}

	return outputFile, outputPath
}

func (t TxtFormatter) provideDefaultWritePath(targetDirectory, relativePath, fileExt string) (string, string) {
	return provideBasePath(targetDirectory, common.ChangeExtension(relativePath, fileExt))
}

func (t TxtFormatter) provideDcpWritePath(targetDirectory, relativePath string) (string, string) {
	return provideBasePath(targetDirectory, relativePath)
}

func (t TxtFormatter) providePartsWritePath(targetDirectory, dirName, fileName string) (string, string) {
	return provideBasePath(targetDirectory, dirName, fileName)
}

func provideBasePath(targetDirectory string, dirParts ...string) (string, string) {
	dirPartsJoined := filepath.Join(dirParts...)
	outputFile := filepath.Join(targetDirectory, dirPartsJoined)
	outputPath := common.GetDir(outputFile)

	return outputFile, outputPath
}
