package integrity

import (
	"bytes"
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/logger"
	"fmt"
	"os"
)

type ILineBreakCounter interface {
	VerifyLineBreaks(partsList components.IList[string], options core.ILockitFileOptions) error
}

type lineBreakCounter struct {
	log logger.ILoggerHandler
}

func NewLineBreakCounter(logger logger.ILoggerHandler) ILineBreakCounter {
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
			errChan <- fmt.Errorf("error when reading file part %s: %w", part, err)
			return
		}

		if err := lc.compareOcorrrences(&data, ocorrencesExpected); err != nil {
			errChan <- fmt.Errorf("error when comparing ocorrences on file part %s: %s", part, err.Error())
			return
		}
	}

	pathList.ForIndex(comparerOcorrencesFunc)

	close(errChan)

	var hasError bool

	for e := range errChan {
		lc.log.LogError(e, "error when comparing line breaks")
		hasError = true
	}

	if hasError {
		return fmt.Errorf("error when comparing line breaks")
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
