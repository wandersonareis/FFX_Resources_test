package verify

import (
	"bytes"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
)

type ILineBreakCounter interface {
	// LockitFile is line break based file for Final Fantasy X game.
	//
	// CountBinaryParts verifies the provided GameData LockitFileParts and counts the line breaks based on the given expected total line breaks for game version.
	// It returns an error if the verification fails.
	//
	// Parameters:
	//   - partsList: A slice of LockitFileParts to be verified and counted.
	//   - options: LockitFileOptions containing the parts sizes and expected line breaks count.
	//
	// Returns:
	//   - error: An error if the verification fails, otherwise nil.
	CountBinaryParts(partsList *[]parts.LockitFileParts, options interactions.LockitFileOptions) error
	CountTextParts(partsList *[]parts.LockitFileParts, options interactions.LockitFileOptions) error
}

type LineBreakCounter struct{}

func (lc LineBreakCounter) CountBinaryParts(partsList *[]parts.LockitFileParts, options interactions.LockitFileOptions) error {
	list := *partsList
	pathList := make([]string, len(list))

	for index, part := range list {
		pathList[index] = part.GetGameData().FullFilePath
	}

	if err := lc.verify(&pathList, options.PartsSizes, len(options.PartsSizes), options.LineBreaksCount); err != nil {
		return fmt.Errorf("error when counting line breaks: %w", err)
	}

	return nil
}

func (lc LineBreakCounter) CountTextParts(partsList *[]parts.LockitFileParts, options interactions.LockitFileOptions) error {
	list := *partsList
	pathList := make([]string, len(list))

	for index, part := range list {
		pathList[index] = part.GetExtractLocation().TargetFile
	}

	if err := lc.verify(&pathList, options.PartsSizes, len(options.PartsSizes), options.LineBreaksCount); err != nil {
		return fmt.Errorf("error when counting line breaks: %w", err)
	}

	return nil
}

func (lc LineBreakCounter) verify(pathList *[]string, ocorrencesCount []int, ocorrencesLength int, expectedLineBreaksCount int) error {
	ocorrencesExpected := 0

	list := *pathList
	for index, part := range list {
		ocorrencesExpected = lc.getOcorrencesExpected(ocorrencesCount, index, ocorrencesLength, expectedLineBreaksCount)

		data, err := lc.readFilePart(part)
		if err != nil {
			return err
		}

		if err := lc.compareOcorrrences(&data, ocorrencesExpected); err != nil {
			return err
		}
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
	case index > 0 && index < ocorrencesLength:
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
