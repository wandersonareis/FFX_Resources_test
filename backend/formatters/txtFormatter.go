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
	targetExtension string
}

func NewTxtFormatter() *TxtFormatter {
	return &TxtFormatter{
		targetExtension: ".txt",
	}
}

func (t TxtFormatter) ReadFile(source interfaces.ISource, targetDirectory string) (string, string) {
	var outputFile, outputPath string

	switch source.Get().Type {
	case models.Folder:
		outputPath = t.provideFolderReadPath(source, targetDirectory)
	case models.Dcp:
		outputFile, outputPath = t.provideDcpReadPath(targetDirectory, source.Get().Name)
	case models.DcpParts:
		outputFile, outputPath = t.providePartsReadPath(targetDirectory, util.DCP_PARTS_TARGET_DIR_NAME, source.Get().Name)
	case models.Lockit:
		outputFile, outputPath = t.provideLockitReadPath(targetDirectory, source.Get().NamePrefix)
	case models.LockitParts:
		outputFile, outputPath = t.providePartsReadPath(targetDirectory, util.LOCKIT_TARGET_DIR_NAME, source.Get().Name)
	default:
		outputFile, outputPath = t.provideDefaulReadPath(targetDirectory, source.Get().RelativePath)
	}

	return outputFile, outputPath
}

func (t TxtFormatter) provideDefaulReadPath(targetDirectory, relativePath string) (string, string) {
	return provideBasePath(targetDirectory, common.ChangeExtension(relativePath, t.targetExtension))
}

func (t *TxtFormatter) provideFolderReadPath(source interfaces.ISource, targetDirectory string) string {
	gameFilesPath := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
	relative := common.MakeRelativePath(gameFilesPath, source.Get().Parent)

	source.Get().RelativePath = relative

	gameVersionDirBase := interactions.NewInteractionService().FFXGameVersion().GetGameVersion().String()

	outputPath := filepath.Join(targetDirectory, gameVersionDirBase, relative)

	return outputPath
}

func (t TxtFormatter) provideDcpReadPath(targetDirectory, fileName string) (string, string) {
	return provideBasePath(targetDirectory, util.DCP_PARTS_TARGET_DIR_NAME, fileName)
}

func (t TxtFormatter) providePartsReadPath(targetDirectory, dirName, fileName string) (string, string) {
	return provideBasePath(targetDirectory, dirName, common.AddExtension(fileName, t.targetExtension))
}

func (t TxtFormatter) provideLockitReadPath(targetDirectory, fileName string) (string, string) {
	return provideBasePath(targetDirectory, util.LOCKIT_TARGET_DIR_NAME, common.AddExtension(fileName, t.targetExtension))
}

func (t TxtFormatter) WriteFile(source interfaces.ISource, targetDirectory string) (string, string) {
	var outputFile, outputPath string

	switch source.Get().Type {
	case models.Dcp:
		outputFile, outputPath = t.provideDcpWritePath(targetDirectory, source.Get().RelativePath)
	case models.DcpParts:
		outputFile, outputPath = t.providePartsWritePath(targetDirectory, util.DCP_PARTS_TARGET_DIR_NAME, source.Get().Name)
	case models.LockitParts:
		outputFile, outputPath = t.providePartsWritePath(targetDirectory, util.LOCKIT_TARGET_DIR_NAME, source.Get().Name)
	default:
		outputFile, outputPath = t.provideDefaultWritePath(targetDirectory, source.Get().RelativePath, source.Get().Extension)
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

	outputFile := filepath.Join(targetDirectory, /* gameVersionDirBase, */ dirPartsJoined)
	outputPath := common.GetDir(outputFile)

	return outputFile, outputPath
}
