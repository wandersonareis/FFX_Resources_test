package lib

import "fmt"

func ErrLockitFilePartsCountMismatch(expectedCount, currentFileCount int) error {
	return fmt.Errorf("error ensuring lockit parts: expected %d, got %d",
		expectedCount, currentFileCount)
}