package components

import (
	"bytes"
	"ffxresources/backend/common"
	"fmt"
	"io"
	"os"
)

type (
	IFileComparer interface {
		CompareFiles() error
	}

	FileComparisonEntry struct {
		FromFile string
		ToFile   string
	}
)

func (fc *FileComparisonEntry) CompareFiles() error {
	common.CheckArgumentNil(fc.FromFile, "compare fromFile")
	common.CheckArgumentNil(fc.ToFile, "compare toFile")

	return fc.compare(fc.FromFile, fc.ToFile)
}

func (fc *FileComparisonEntry) compare(fromFile, toFile string) error {
	if err := fc.checkFileSizes(fromFile, toFile); err != nil {
		return err
	}

	fromF, toF, err := fc.openFiles(fromFile, toFile)
	if err != nil {
		return err
	}
	defer fromF.Close()
	defer toF.Close()

	if err := fc.compareFilesContent(fromF, toF); err != nil {
		return err
	}

	return nil
}

func (fc *FileComparisonEntry) checkFileSizes(fromFile, toFile string) error {
	fromInfo, err := os.Stat(fromFile)
	if err != nil {
		return fmt.Errorf("error reading file info for '%s': %v", fromFile, err)
	}

	toInfo, err := os.Stat(toFile)
	if err != nil {
		return fmt.Errorf("error reading file info for '%s': %v", toFile, err)
	}

	if fromInfo.Size() != toInfo.Size() {
		return fmt.Errorf("size mismatch detected: '%s' is %d bytes, '%s' is %d bytes",
			fromFile, fromInfo.Size(), toFile, toInfo.Size())
	}

	return nil
}

func (fc *FileComparisonEntry) openFiles(fromFile, toFile string) (*os.File, *os.File, error) {
	fromF, err := os.Open(fromFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening file %s: %v", fromFile, err)
	}

	toF, err := os.Open(toFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening file %s: %v", toFile, err)
	}

	return fromF, toF, nil
}

func (fc *FileComparisonEntry) compareFilesContent(fromF, toF *os.File) error {
	const bufferSize = 64 * 1024 // 64KB buffer
	fromBuf := make([]byte, bufferSize)
	toBuf := make([]byte, bufferSize)

	for {
		fromN, fromErr := fromF.Read(fromBuf)
		toN, toErr := toF.Read(toBuf)

		if fromErr != nil && fromErr != io.EOF {
			return fmt.Errorf("error reading file '%s': %v", fc.FromFile, fromErr)
		}
		if toErr != nil && toErr != io.EOF {
			return fmt.Errorf("error reading file '%s': %v", fc.ToFile, toErr)
		}

		if fromN != toN {
			return fmt.Errorf("size mismatch detected between '%s' and '%s'", fc.FromFile, fc.ToFile)
		}

		if fromErr == io.EOF && toErr == io.EOF {
			break
		}

		if !bytes.Equal(fromBuf[:fromN], toBuf[:toN]) {
			return fmt.Errorf("content mismatch detected between '%s' and '%s'", fc.FromFile, fc.ToFile)
		}
	}

	return nil
}
