package spira

import (
	"ffxresources/backend/formats"
	"ffxresources/backend/lib"
)

func NewFileProcessor(fileInfo lib.FileInfo) lib.IFileProcessor {
	fileType := fileInfo.Type

	interaction := lib.NewInteraction()

	extractLocation := interaction.ExtractLocation

	extractPath, err := extractLocation.ProvideTargetDirectory()
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	translateLocation := lib.NewInteraction().TranslateLocation

	translatePath, err := translateLocation.ProvideTargetDirectory()
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	reimportLocation := lib.NewInteraction().ImportLocation

	_, err = reimportLocation.ProvideTargetDirectory()
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	provideLocationsToFileInfo(&fileInfo)

	switch fileType {
	case lib.Dialogs, lib.Tutorial, lib.DcpParts:
		return formats.NewDialogs(fileInfo)
	case lib.Kernel:
		return formats.NewKernel(fileInfo)
	case lib.Dcp:
		return formats.NewDcpFile(fileInfo)
	case lib.Folder:
		return NewSpiraFolder(fileInfo, extractPath, translatePath)
	default:
		return nil
	}
}

func provideLocationsToFileInfo(fileInfo *lib.FileInfo) {
	extractLocation := lib.NewInteraction().ExtractLocation	

	translateLocation := lib.NewInteraction().TranslateLocation	

	reimportLocation := lib.NewInteraction().ImportLocation	

	fileInfo.ExtractLocation = *extractLocation
	fileInfo.TranslateLocation = *translateLocation
	fileInfo.ImportLocation = *reimportLocation
}