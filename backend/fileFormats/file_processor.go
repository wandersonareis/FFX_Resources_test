package fileFormats

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/dcp"
	"ffxresources/backend/fileFormats/internal/folder"
	"ffxresources/backend/fileFormats/internal/lockit"
	"ffxresources/backend/fileFormats/internal/text"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
)

// formats is a map that associates models.NodeType values with functions that
// create instances of interactions.IFileProcessor. Each entry in the map
// corresponds to a specific type of node
var formats = map[models.NodeType]func(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor{
	models.Dialogs:        text.NewDialogs,
	models.DialogsSpecial: text.NewDialogs,
	models.Tutorial:       text.NewDialogs,
	models.DcpParts:       text.NewDialogs,
	models.Kernel:         text.NewKernel,
	models.Dcp:            dcp.NewDcpFile,
	models.Lockit:         lockit.NewLockitFile,
	models.Folder:         folder.NewSpiraFolder,
}

func NewFileExtractor(source interfaces.ISource, destination locations.IDestination) interfaces.IExtractor {
	return NewFileProcessor(source, destination)
}

func NewFileCompressor(source interfaces.ISource, destination locations.IDestination) interfaces.ICompressor {
	return NewFileProcessor(source, destination)
}

func NewFileProcessor(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	fileType := source.GetType()

	if value, ok := formats[fileType]; ok {
		return value(source, destination)
	}

	return nil
}
