package lib

func EnsurePartsListCount(expectedCount, currentCount int) error {
	if expectedCount != currentCount {
		return ErrDcpFilePartsCountMismatch(expectedCount, currentCount)
	}

	return nil
}