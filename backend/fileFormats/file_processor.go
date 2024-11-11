package fileFormats

import (
	"ffxresources/backend/events"
	"ffxresources/backend/fileFormats/internal/dcp"
	"ffxresources/backend/fileFormats/internal/dlg"
	"ffxresources/backend/fileFormats/internal/lockit"
	"ffxresources/backend/fileFormats/internal/mt2"
	"ffxresources/backend/interactions"
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

	switch fileType {
	case models.Dialogs, models.Tutorial, models.DcpParts:
		return dlg.NewDialogs(dataInfo)
	case models.Kernel:
		return mt2.NewKernel(dataInfo)
	case models.Dcp:
		return dcp.NewDcpFile(dataInfo)
	case models.Lockit:
		return lockit.NewLockitFile(dataInfo)
	case models.Folder:
		extractPath, err := interactions.NewInteraction().ExtractLocation.ProvideTargetDirectory()
		if err != nil {
			events.NotifyError(err)
			return nil
		}

		translatePath, err := interactions.NewInteraction().TranslateLocation.ProvideTargetDirectory()
		if err != nil {
			events.NotifyError(err)
			return nil
		}

		return NewSpiraFolder(dataInfo, extractPath, translatePath)
	default:
		return nil
	}
}
