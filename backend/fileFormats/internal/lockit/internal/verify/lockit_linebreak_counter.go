package verify

import (
	"bytes"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitFileParts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"fmt"
	"os"
)

type ILineBreakCounter interface {
	// CountBinaryParts verifies the provided GameData LockitFileParts and counts the line breaks based on the given expected total line breaks for game version.
	// It returns an error if the verification fails.
	//
	// LockitFile is line break based file for Final Fantasy X game.
	//
	// Parameters:
	//   - partsList: A slice of LockitFileParts to be verified and counted.
	//   - options: LockitFileOptions containing the parts sizes and expected line breaks count.
	//
	// Returns:
	//   - error: An error if the verification fails, otherwise nil.
	CountBinaryParts(partsList components.IList[lockitFileParts.LockitFileParts], options interactions.LockitFileOptions) error
	CountTextParts(partsList components.IList[lockitFileParts.LockitFileParts], options interactions.LockitFileOptions) error
}

type LineBreakCounter struct{}

func (lc LineBreakCounter) CountBinaryParts(partsList components.IList[lockitFileParts.LockitFileParts], options interactions.LockitFileOptions) error {
	pathList := components.NewList[string](partsList.GetLength())

	partsList.ForEach(func(part lockitFileParts.LockitFileParts) {
		pathList.Add(part.Source().Get().Path)
	})

	if err := lc.verify(pathList, options.PartsSizes, len(options.PartsSizes), options.LineBreaksCount); err != nil {
		return fmt.Errorf("error when counting line breaks: %w", err)
	}

	return nil
}

func (lc LineBreakCounter) CountTextParts(partsList components.IList[lockitFileParts.LockitFileParts], options interactions.LockitFileOptions) error {
	pathList := components.NewEmptyList[string]()

	partsList.ForEach(func(part lockitFileParts.LockitFileParts) {
		pathList.Add(part.Destination().Extract().Get().GetTargetFile())
	})

	if err := lc.verify(pathList, options.PartsSizes, len(options.PartsSizes), options.LineBreaksCount); err != nil {
		return fmt.Errorf("error when counting line breaks: %w", err)
	}

	return nil
}

func (lc LineBreakCounter) verify(pathList components.IList[string], ocorrencesCount []int, ocorrencesLength int, expectedLineBreaksCount int) error {
	errChan := make(chan error, pathList.GetLength())
	successChan := make(chan string, pathList.GetLength())

	defer close(errChan)
	defer close(successChan)

	loggerHandler := &logger.LogHandler{
		Logger: logger.Get().With().Str("module", "linebreak_counter").Logger(),
	}

	comparerOcorrencesFunc := func(index int, part string) {
		ocorrencesExpected := lc.getOcorrencesExpected(ocorrencesCount, index, ocorrencesLength, expectedLineBreaksCount)

		data, err := lc.readFilePart(part)
		if err != nil {
			errChan <- fmt.Errorf("error when reading file part %s: %w", part, err)
			return
		}

		if err := lc.compareOcorrrences(&data, ocorrencesExpected); err != nil {
			errChan <- fmt.Errorf("error when comparing ocorrences on file %s: %w", part, err)
			return
		}
		successChan <- part
	}

	pathList.ForIndex(comparerOcorrencesFunc)

	select {
	case err := <-errChan:
		loggerHandler.LogError(err, "error when comparing line breaks")
	case <-successChan:
		loggerHandler.LogInfo("line breaks comparison successfully for: %s", <-successChan)
	}

	return nil
}

func (lc LineBreakCounter) readFilePart(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error when reading file part %s", path)
	}

	return data, nil
}

func (lc LineBreakCounter) getOcorrencesExpected(ocorrencesCount []int, index, ocorrencesLength, expectedLineBreaksCount int) int {
	ocorrencesExpected := 0

	switch true {
	case index == 0:
		ocorrencesExpected = ocorrencesCount[index]
	case index < ocorrencesLength:
		ocorrencesExpected = ocorrencesCount[index] - ocorrencesCount[index-1]
	case index <= ocorrencesLength:
		ocorrencesExpected = expectedLineBreaksCount - ocorrencesCount[index-1]
	}
	return ocorrencesExpected
}

func (lc LineBreakCounter) countLineBreaks(data *[]byte) int {
	return bytes.Count(*data, []byte{0x0d, 0x0a})
}

func (lc LineBreakCounter) compareOcorrrences(data *[]byte, expected int) error {
	ocorrences := lc.countLineBreaks(data)

	if ocorrences != expected {
		return fmt.Errorf("file has %d line breaks, expected %d", ocorrences, expected)
	}

	return nil
}
