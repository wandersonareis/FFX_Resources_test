package locationsBase

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interfaces"
	"path/filepath"
)

type (
	IBuildTargets interface {
		BuildExtractOutput(source interfaces.ISource, formatter interfaces.ITextFormatter) error
		BuildImportOutput(source interfaces.ISource, formatter interfaces.ITextFormatter) error
	}
	IOperationStrategy interface {
		Process(lb *LocationBase, source interfaces.ISource, formatter interfaces.ITextFormatter)
	}
)

type readOperation struct{}

func (r readOperation) Process(lb *LocationBase, source interfaces.ISource, formatter interfaces.ITextFormatter) {
	lb.TargetFile, lb.TargetPath = formatter.ReadFile(source, lb.GetTargetDirectory())
}

type writeOperation struct{}

func (w writeOperation) Process(lb *LocationBase, source interfaces.ISource, formatter interfaces.ITextFormatter) {
	lb.TargetFile, lb.TargetPath = formatter.WriteFile(source, lb.GetTargetDirectory())
}

func (lb *LocationBase) buildTargetOutput(
	strategy IOperationStrategy,
	source interfaces.ISource,
	formatter interfaces.ITextFormatter,
) error {
	if err := common.CheckArgumentNil(source, "source"); err != nil {
		return err
	}

	if err := common.CheckArgumentNil(formatter, "formatter"); err != nil {
		return err
	}

	strategy.Process(lb, source, formatter)

	if !source.IsDir() && common.IsValidFilePath(lb.TargetFile) {
		lb.TargetFileName = filepath.Base(lb.TargetFile)
	}

	lb.IsExist = lb.IsTargetFileAvailable()

	return nil
}

func (lb *LocationBase) BuildExtractOutput(
	source interfaces.ISource,
	formatter interfaces.ITextFormatter,
) error {
	if err := lb.buildTargetOutput(readOperation{}, source, formatter); err != nil {
		return err
	}
	return nil
}

func (lb *LocationBase) BuildImportOutput(
	source interfaces.ISource,
	formatter interfaces.ITextFormatter,
) error {
	if err := lb.buildTargetOutput(writeOperation{}, source, formatter); err != nil {
		return err
	}
	return nil
}
