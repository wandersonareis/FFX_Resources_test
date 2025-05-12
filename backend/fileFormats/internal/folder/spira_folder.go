package folder

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/baseFormats"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"fmt"
)

type SpiraFolder struct {
	baseFormats.IBaseFileFormat
	loggingService.ILoggerService
}

func NewSpiraFolder(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	return &SpiraFolder{
		IBaseFileFormat: baseFormats.NewFormatsBase(source, destination),

		ILoggerService: &loggingService.LoggerService{
			Logger: loggingService.Get().With().Str("module", "spira_folder").Logger(),
		},
	}
}

func (sf *SpiraFolder) Extract() error {
	return fmt.Errorf("use DirectoryExtractService instead")
}

func (sf *SpiraFolder) Compress() error {
	return fmt.Errorf("use DirectoryCompressService instead")
}
