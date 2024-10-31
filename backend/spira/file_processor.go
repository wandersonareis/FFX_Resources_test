package spira

import (
	"ffxresources/backend/formats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"ffxresources/backend/models"
)

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
		return formats.NewDialogs(dataInfo)
	case models.Kernel:
		return formats.NewKernel(dataInfo)
	case models.Dcp:
		return formats.NewDcpFile(dataInfo)
	case models.Lockit:
		return formats.NewLockitFile(dataInfo)
	case models.Folder:
		return NewSpiraFolder(dataInfo, extractPath, translatePath)
	default:
		return nil
	}
}
