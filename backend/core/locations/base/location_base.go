package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"os"
)

type (
	ILocationBase interface {
		ITargetDirectoryBase
		ITargetFileBase

		BuildTargetReadOutput(source interfaces.ISource, formatter interfaces.ITextFormatter)
		BuildTargetWriteOutput(source interfaces.ISource, formatter interfaces.ITextFormatter)
		Dispose()
	}
	LocationBase struct {
		TargetDirectoryBase
		TargetFileBase
	}
)

func NewLocationBase(options *LocationBaseOptions) LocationBase {
	return LocationBase{
		TargetDirectoryBase: TargetDirectoryBase{
			TargetDirectory:     options.TargetDirectory,
			TargetDirectoryName: options.TargetDirectoryName,
		},
		TargetFileBase: *new(TargetFileBase),
	}
}

func (lb *LocationBase) Dispose() {
	if common.IsFileExists(lb.GetTargetFile()) {
		err := os.Remove(lb.TargetFile)
		if err != nil {
			l := logger.Get()
			l.Error().Msgf("error when removing file: %s", err)
		}
	}
}
