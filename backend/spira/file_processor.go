package spira

import (
	"ffxresources/backend/common"
	"ffxresources/backend/formats"
	"ffxresources/backend/lib"
)

func NewFileProcessor(fileInfo *lib.FileInfo) lib.IFileProcessor {
	fileType := fileInfo.Type

	provideLocationsToFileInfo(fileInfo)
	
	extractPath, err := lib.NewInteraction().ExtractLocation.ProvideTargetDirectory()
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	translatePath, err := lib.NewInteraction().TranslateLocation.ProvideTargetDirectory()
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	_, err = lib.NewInteraction().ImportLocation.ProvideTargetDirectory()
	if err != nil {
		lib.NotifyError(err)
		return nil
	}


	switch fileType {
	case common.Dialogs, common.Tutorial, common.DcpParts:
		return formats.NewDialogs(fileInfo)
	case common.Kernel:
		return formats.NewKernel(fileInfo)
	case common.Dcp:
		return formats.NewDcpFile(fileInfo)
	case common.Folder:
		return NewSpiraFolder(fileInfo, extractPath, translatePath)
	default:
		return nil
	}
}

func provideLocationsToFileInfo(fileInfo *lib.FileInfo) {
	fileInfo.ExtractLocation = *lib.NewInteraction().ExtractLocation
	fileInfo.TranslateLocation = *lib.NewInteraction().TranslateLocation
	fileInfo.ImportLocation = *lib.NewInteraction().ImportLocation
}