package integrity

import (
	"bytes"
	"ffxresources/backend/core/components"
	"ffxresources/backend/loggingService"
	"ffxresources/backend/models"
	"fmt"
	"io"
	"os"
)

type (
	IComparerContent interface {
		CompareContent(filesList components.IList[models.FileComparisonEntry]) error
	}

	comparerContent struct {
		log loggingService.ILoggerService
	}
)

// TODO: Review the implementation of this function
func NewComparerContent(loggerHandler loggingService.ILoggerService) IComparerContent {
	return &comparerContent{
		log: loggerHandler,
	}
}

func (pc *comparerContent) CompareContent(filesList components.IList[models.FileComparisonEntry]) error {
	errChan := make(chan error, filesList.GetLength())

	filesList.ParallelForEach(func(item models.FileComparisonEntry) {
		if err := pc.compare(item.FromFile, item.ToFile); err != nil {
			errChan <- err
			return
		}
	})

	close(errChan)

	var hasError bool
	for e := range errChan {
		pc.log.Error(e, "error when comparing text content")
		hasError = true
	}

	if hasError {
		return fmt.Errorf("errors occurred during comparison")
	}

	return nil
}

func (pc *comparerContent) compare(fromFile, toFile string) error {
	if err := pc.checkFileSizes(fromFile, toFile); err != nil {
		return err
	}

	fromF, toF, err := pc.openFiles(fromFile, toFile)
	if err != nil {
		return err
	}
	defer fromF.Close()
	defer toF.Close()

	if err := pc.compareFilesContent(fromF, toF); err != nil {
		return err
	}

	return nil
}

func (pc *comparerContent) checkFileSizes(fromFile, toFile string) error {
	readingErrorMsg := "error reading file info"
	fromInfo, err := os.Stat(fromFile)
	if err != nil {
		pc.log.Error(err, readingErrorMsg, fromFile)
		return fmt.Errorf("%s", readingErrorMsg)
	}

	toInfo, err := os.Stat(toFile)
	if err != nil {
		pc.log.Error(err, readingErrorMsg, toFile)
		return fmt.Errorf("%s", readingErrorMsg)
	}

	if fromInfo.Size() != toInfo.Size() {
		pc.log.Error(nil, "size mismatch detected between %s and %s", fromFile, toFile)
		return fmt.Errorf("size mismatch detected between files")
	}

	return nil
}

func (pc *comparerContent) openFiles(fromFile, toFile string) (*os.File, *os.File, error) {
	openingErrorMsg := "error opening file"
	fromF, err := os.Open(fromFile)
	if err != nil {
		pc.log.Error(err, openingErrorMsg, fromFile)
		return nil, nil, fmt.Errorf("%s", openingErrorMsg)
	}

	toF, err := os.Open(toFile)
	if err != nil {
		pc.log.Error(err, openingErrorMsg, toFile)
		return nil, nil, fmt.Errorf("%s", openingErrorMsg)
	}

	return fromF, toF, nil
}

func (pc *comparerContent) compareFilesContent(fromF, toF *os.File) error {
	const bufferSize = 64 * 1024 // 64KB buffer
	fromBuf := make([]byte, bufferSize)
	toBuf := make([]byte, bufferSize)

	openFileErrorMsg := "error reading file"
	mismatchErrMsg := "content mismatch detected between files"

	for {
		fromN, fromErr := fromF.Read(fromBuf)
		toN, toErr := toF.Read(toBuf)

		if fromErr != nil && fromErr != io.EOF {
			pc.log.Error(fromErr, openFileErrorMsg)
			return fmt.Errorf("%s", openFileErrorMsg)
		}
		if toErr != nil && toErr != io.EOF {
			pc.log.Error(toErr, openFileErrorMsg)
			return fmt.Errorf("%s", openFileErrorMsg)
		}

		if fromN != toN {
			pc.log.Error(nil, mismatchErrMsg)
			return fmt.Errorf("%s", mismatchErrMsg)
		}

		if fromErr == io.EOF && toErr == io.EOF {
			break
		}

		if !bytes.Equal(fromBuf[:fromN], toBuf[:toN]) {
			pc.log.Error(nil, mismatchErrMsg)
			return fmt.Errorf("%s", mismatchErrMsg)
		}
	}

	return nil
}
