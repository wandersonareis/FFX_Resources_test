package integrity

import (
	"bytes"
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/loggingService"
	"fmt"
	"os"
)

type ILineBreakCounter interface {
	VerifyLineBreaks(partsList components.IList[string], options core.ILockitFileOptions) error
}

type lineBreakCounter struct {
	log loggingService.ILoggerService
}

func NewLineBreakCounter(logger loggingService.ILoggerService) ILineBreakCounter {
	return &lineBreakCounter{log: logger}
}

func (lc *lineBreakCounter) VerifyLineBreaks(partsList components.IList[string], options core.ILockitFileOptions) error {
	if err := lc.verify(partsList, options.GetPartsSizes()); err != nil {
		return fmt.Errorf("error when counting line breaks: %w", err)
	}

	return nil
}

func (lc *lineBreakCounter) verify(pathList components.IList[string], partsSizes []int) error {
	errChan := make(chan error, pathList.GetLength())

	comparerOcorrencesFunc := func(index int, part string) {
		ocorrencesExpected := partsSizes[index]

		data, err := lc.readFilePart(part)
		if err != nil {
			lc.log.Error(err, "error when reading file part %s", part)
			errChan <- err
			return
		}

		if err := lc.compareOcorrrences(&data, ocorrencesExpected); err != nil {
			lc.log.Error(err, "error when comparing ocorrences on file part %s", part)
			errChan <- err
			return
		}
	}

	pathList.ForIndex(comparerOcorrencesFunc)

	close(errChan)

	for e := range errChan {
		if e != nil {
			return e
		}
	}

	return nil
}

func (lc lineBreakCounter) readFilePart(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error when reading file part %s", path)
	}

	return data, nil
}

func (lc *lineBreakCounter) countLineBreaks(data *[]byte) int {
	return bytes.Count(*data, []byte{0x0d, 0x0a})
}

func (lc *lineBreakCounter) compareOcorrrences(data *[]byte, expected int) error {
	ocorrences := lc.countLineBreaks(data)

	if ocorrences != expected {
		return fmt.Errorf("the file has %d line breaks, expected %d", ocorrences, expected)
	}

	return nil
}
