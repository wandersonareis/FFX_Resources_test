package lib

import "fmt"

func ErrDcpFileExtractIntegrityFailed(err error) error {
	return fmt.Errorf("error when checking dcp file extracted integrity: %v", err)
}

func ErrDcpFileFailedToExtract() error {
	return fmt.Errorf("failed to extract DCP file")
}

func ErrDcpFilePartsCountMismatch(expected, actual int) error {
	return fmt.Errorf("dcp file parts count mismatch: expected %d, got %d", expected, actual)
}

func ErrFailedToCompressFileParts() error {
	return fmt.Errorf("failed to compress file parts")
}

func ErrInvalidFileSize(path string) error {
	return fmt.Errorf("invalid file size: %s", path)
}
