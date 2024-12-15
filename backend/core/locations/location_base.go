package locations

import (
	"ffxresources/backend/bases"
	"ffxresources/backend/common"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"os"
	"path/filepath"
)

type LocationBase struct {
	bases.TargetDirectoryBase
	bases.TargetFileBase
}

func NewLocationBase(options *bases.LocationBaseOptions) LocationBase {
	return LocationBase{
		TargetDirectoryBase: bases.TargetDirectoryBase{
			TargetDirectory:     options.TargetDirectory,
			TargetDirectoryName: options.TargetDirectoryName,
		},
		TargetFileBase:      *new(bases.TargetFileBase),
	}
}

func (lb *LocationBase) BuildTargetOutput(source interfaces.ISource, formatter interfaces.ITextFormatterDev) {
	lb.TargetFile, lb.TargetPath = formatter.ReadFile(source, lb.TargetDirectory)

	if !source.Get().IsDir {
		lb.TargetFileName = filepath.Base(lb.TargetFile)
	}

	lb.IsExist = lb.IsTargetFileAvailable()
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
