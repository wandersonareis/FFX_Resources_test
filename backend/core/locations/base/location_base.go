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

		BuildTargetReadOutput(source interfaces.ISource, formatter interfaces.ITextFormatterDev)
		BuildTargetWriteOutput(source interfaces.ISource, formatter interfaces.ITextFormatterDev)
		DisposeTargetFile()
		DisposeTargetPath()
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

func (lb *LocationBase) DisposeTargetFile() {
	if common.IsFileExists(lb.GetTargetFile()) {
		err := os.Remove(lb.TargetFile)
		if err != nil {
			l := logger.Get()
			l.Error().Msgf("error when removing file: %s", err)
		}
	}
}

func (lb *LocationBase) DisposeTargetPath() {
	if common.IsPathExists(lb.TargetPath) {
		err := os.RemoveAll(lb.TargetPath)
		if err != nil {
			l := logger.Get()
			l.Error().Msgf("error when removing path: %s", err)
			return
		}
	}
}
