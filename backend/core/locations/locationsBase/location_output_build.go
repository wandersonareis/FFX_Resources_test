package locationsBase

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interfaces"
	"path/filepath"
)

type BuildOperationType int

const (
	opRead BuildOperationType = iota
	opWrite
)

func (lb *LocationBase) buildTargetOutput(
	source interfaces.ISource,
	formatter interfaces.ITextFormatter,
	op BuildOperationType,
) {
	common.CheckArgumentNil(source, "source")
	if formatter == nil {
		panic("formatter cannot be nil")
	}

	if op == opRead {
		lb.TargetFile, lb.TargetPath = formatter.ReadFile(source, lb.TargetDirectory)
	}

	if op == opWrite {
		lb.TargetFile, lb.TargetPath = formatter.WriteFile(source, lb.TargetDirectory)
	}

	if !source.Get().IsDir && common.IsValidFilePath(lb.TargetFile) {
		lb.TargetFileName = filepath.Base(lb.TargetFile)
	}
	lb.IsExist = lb.IsTargetFileAvailable()
}

func (lb *LocationBase) BuildTargetReadOutput(
	source interfaces.ISource,
	formatter interfaces.ITextFormatter,
) {
	lb.buildTargetOutput(source, formatter, opRead)
}

func (lb *LocationBase) BuildTargetWriteOutput(
	source interfaces.ISource,
	formatter interfaces.ITextFormatter,
) {
	lb.buildTargetOutput(source, formatter, opWrite)
}
