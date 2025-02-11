package common

import (
	"ffxresources/backend/logger"
	"fmt"
)

func RecoverFn(fn func() error) error {
	l := logger.NewLoggerHandler("panic_recover")
	logger := l.GetLogger().Logger.Fatal()
	var outError error
	defer func() {
		if r := recover(); r != nil {
			logger.Msgf("Recovered from panic: %v", r)
			outError = fmt.Errorf("a fatal error occurred while processing the request")
		}
	}()

	outError = fn()

	return outError
}
