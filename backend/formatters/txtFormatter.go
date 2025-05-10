package formatters

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"path/filepath"
)

type TxtFormatter struct {
	TargetExtension string
	GameVersionDir  string
	GameFilesPath   string
}

func NewTxtFormatter() *TxtFormatter {
	return &TxtFormatter{
		TargetExtension: ".txt",
		GameVersionDir:  interactions.NewInteractionService().FFXGameVersion().GetGameVersion().String(),
		GameFilesPath:   interactions.NewInteractionService().GameLocation.GetTargetDirectory(),
	}
}

func (t TxtFormatter) ReadFile(source interfaces.ISource, targetDirectory string) (string, string) {
	var outputFile, outputPath string

	switch source.GetType() {
	case models.Folder:
		outputPath = t.provideFolderReadPath(source, targetDirectory)
	case models.Dcp:
		outputFile, outputPath = t.provideDcpReadPath(targetDirectory, source.GetName())
	case models.DcpParts:
		outputFile, outputPath = t.providePartsReadPath(targetDirectory, util.DCP_PARTS_TARGET_DIR_NAME, source.GetName())
	case models.Lockit:
		outputFile, outputPath = t.provideLockitReadPath(targetDirectory, source.GetNameWithoutExtension())
	case models.LockitParts:
		outputFile, outputPath = t.providePartsReadPath(targetDirectory, util.LOCKIT_TARGET_DIR_NAME, source.GetName())
	default:
		outputFile, outputPath = t.provideDefaulReadPath(targetDirectory, source.GetRelativePath())
		if !common.IsValidFilePath(outputFile) {
			outputFile = ""
		}
	}

	return outputFile, outputPath
}

func (t TxtFormatter) provideDefaulReadPath(targetDirectory, relativePath string) (string, string) {
	return provideBasePath(targetDirectory, common.ChangeExtension(relativePath, t.TargetExtension))
}

func (t *TxtFormatter) provideFolderReadPath(source interfaces.ISource, targetDirectory string) string {
	relative := common.MakeRelativePath(t.GameFilesPath, source.GetParentPath())
	source.SetRelativePath(relative)

	outputPath := filepath.Join(targetDirectory, t.GameVersionDir, relative)

	return outputPath
}

func (t TxtFormatter) provideDcpReadPath(targetDirectory, fileName string) (string, string) {
	return provideBasePath(targetDirectory, util.DCP_PARTS_TARGET_DIR_NAME, fileName)
}

func (t TxtFormatter) providePartsReadPath(targetDirectory, dirName, fileName string) (string, string) {
	return provideBasePath(targetDirectory, dirName, common.AddExtension(fileName, t.TargetExtension))
}

func (t TxtFormatter) provideLockitReadPath(targetDirectory, fileName string) (string, string) {
	return provideBasePath(targetDirectory, util.LOCKIT_TARGET_DIR_NAME, common.AddExtension(fileName, t.TargetExtension))
}

func (t TxtFormatter) WriteFile(source interfaces.ISource, targetDirectory string) (string, string) {
	var outputFile, outputPath string

	switch source.GetType() {
	case models.Dcp:
		outputFile, outputPath = t.provideDcpWritePath(targetDirectory, source.GetRelativePath())
	case models.DcpParts:
		outputFile, outputPath = t.providePartsWritePath(targetDirectory, util.DCP_PARTS_TARGET_DIR_NAME, source.GetName())
	case models.LockitParts:
		outputFile, outputPath = t.providePartsWritePath(targetDirectory, util.LOCKIT_TARGET_DIR_NAME, source.GetName())
	default:
		outputFile, outputPath = t.provideDefaultWritePath(targetDirectory, source.GetRelativePath(), source.Get().Extension)
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
