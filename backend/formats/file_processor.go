package formats

import (
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"ffxresources/backend/models"
)

func NewFileExtractor(dataInfo *interactions.GameDataInfo) models.IExtractor {
	return NewFileProcessor(dataInfo)
}

func NewFileCompressor(dataInfo *interactions.GameDataInfo) models.ICompressor {
	return NewFileProcessor(dataInfo)
}

func NewFileProcessor(dataInfo *interactions.GameDataInfo) interactions.IFileProcessor {
	fileType := dataInfo.GameData.Type

	extractPath, err := interactions.NewInteraction().ExtractLocation.ProvideTargetDirectory()
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	translatePath, err := interactions.NewInteraction().TranslateLocation.ProvideTargetDirectory()
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	_, err = interactions.NewInteraction().ImportLocation.ProvideTargetDirectory()
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	switch fileType {
	case models.Dialogs, models.Tutorial, models.DcpParts:
		return NewDialogs(dataInfo)
	case models.Kernel:
		return NewKernel(dataInfo)
	case models.Dcp:
		return NewDcpFile(dataInfo)
	case models.Lockit:
		return NewLockitFile(dataInfo)
	case models.Folder:
		return NewSpiraFolder(dataInfo, extractPath, translatePath)
	default:
		return nil
	}
}
