package fileFormats

import (
	"ffxresources/backend/events"
	"ffxresources/backend/fileFormats/internal/dcp"
	"ffxresources/backend/fileFormats/internal/folder"
	"ffxresources/backend/fileFormats/internal/lockit"
	"ffxresources/backend/fileFormats/internal/text/dlg"
	"ffxresources/backend/fileFormats/internal/text/mt2"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
)

// formats is a map that associates models.NodeType values with functions that
// create instances of interactions.IFileProcessor. Each entry in the map
// corresponds to a specific type of node
var formats = map[models.NodeType]func(interactions.IGameDataInfo) interactions.IFileProcessor{
	models.Dialogs:        dlg.NewDialogs,
	models.DialogsSpecial: dlg.NewDialogs,
	models.Tutorial:       dlg.NewDialogs,
	models.DcpParts:       dlg.NewDialogs,
	models.Kernel:         mt2.NewKernel,
	models.Dcp:            dcp.NewDcpFile,
	models.Lockit:         lockit.NewLockitFile,
}

func NewFileExtractor(dataInfo interactions.IGameDataInfo) models.IExtractor {
	return NewFileProcessor(dataInfo)
}

func NewFileCompressor(dataInfo interactions.IGameDataInfo) models.ICompressor {
	return NewFileProcessor(dataInfo)
}

func NewFileProcessor(dataInfo interactions.IGameDataInfo) interactions.IFileProcessor {
	fileType := dataInfo.GetGameData().Type

	if fileType == models.Folder {
		return folder.NewSpiraFolder(dataInfo, NewFileProcessor)
	}

	if err := interactions.NewInteraction().ExtractLocation.ProvideTargetDirectory(); err != nil {
		events.NotifyError(err)
		return nil
	}

	if err := interactions.NewInteraction().TranslateLocation.ProvideTargetDirectory(); err != nil {
		events.NotifyError(err)
		return nil
	}

	if value, ok := formats[fileType]; ok {
		return value(dataInfo)
	}

	return nil
}
