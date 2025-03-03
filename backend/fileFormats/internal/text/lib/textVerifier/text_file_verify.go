package textVerifier

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
	"os"
	"sync"
)

type TextIntegrityVerifyFunc func(source interfaces.ISource, destination locations.IDestination, segmentCounter ISegmentCounter, fileComparer IComparer, logger logger.ILoggerHandler) error

var (
	ExtractIntegrityCheck  TextIntegrityVerifyFunc = extractIntegrityCheck
	CompressIntegrityCheck TextIntegrityVerifyFunc = compressIntegrityCheck
)

type (
	ITextVerifier interface {
		Verify(interfaces.ISource, locations.IDestination, TextIntegrityVerifyFunc) error
	}

	TextVerifier struct {
		fileSegmentCounter  ISegmentCounter
		fileContentComparer IComparer

		initializeSegmentCounter sync.Once
		initializeFilesComparer  sync.Once

		log logger.ILoggerHandler
	}
)

func NewTextsVerify() *TextVerifier {
	return &TextVerifier{
		fileContentComparer: newPartComparer(),

		log: &logger.LogHandler{
			Logger: logger.Get().With().Str("module", "texts_verify").Logger(),
		},
	}
}

func (dv *TextVerifier) Verify(source interfaces.ISource, destination locations.IDestination, verify TextIntegrityVerifyFunc) error {
	extractLocation := destination.Extract().Get()
	if err := extractLocation.Validate(); err != nil {
		return err
	}

	dv.initializeSegmentCounter.Do(func() {
		dv.fileSegmentCounter = new(segmentCounter)
	})

	dv.initializeFilesComparer.Do(func() {
		dv.fileContentComparer = newPartComparer()
	})

	if err := verify(source, destination, dv.fileSegmentCounter, dv.fileContentComparer, dv.log); err != nil {
		return err
	}

	return nil
}

func extractIntegrityCheck(source interfaces.ISource, destination locations.IDestination, segmentCounter ISegmentCounter, fileComparer IComparer, logger logger.ILoggerHandler) error {
	extractLocation := destination.Extract().Get()

	if err := segmentCounter.CountBinary(extractLocation.GetTargetFile()); err != nil {
		if err := os.Remove(extractLocation.GetTargetFile()); err != nil {
			logger.LogError(err, "failed to remove broken text file: %s", extractLocation.GetTargetFile())
			return fmt.Errorf("failed to integrity text file: %s", extractLocation.GetTargetFile())
		}

		return err
	}

	if err := segmentCounter.CountText(extractLocation.GetTargetFile()); err != nil {
		if err := os.Remove(extractLocation.GetTargetFile()); err != nil {
			logger.LogError(err, "failed to remove broken text file: %s", extractLocation.GetTargetFile())
			return fmt.Errorf("failed to integrity text file: %s", extractLocation.GetTargetFile())
		}

		return err
	}

	return nil
}

func compressIntegrityCheck(source interfaces.ISource, destination locations.IDestination, segmentCounter ISegmentCounter, fileComparer IComparer, logger logger.ILoggerHandler) error {
	extractLocation := destination.Extract().Get()
	translateLocation := destination.Translate().Get()
	importLocation := destination.Import().Get()

	if err := importLocation.Validate(); err != nil {
		return err
	}

	if err := fileComparer.CompareTranslatedTextParts(translateLocation.GetTargetFile(), extractLocation.GetTargetFile()); err != nil {
		if err := os.Remove(importLocation.GetTargetFile()); err != nil {
			logger.LogError(err, "failed to remove broken text file: %s", importLocation.GetTargetFile())
			return fmt.Errorf("failed to integrity text file: %s", importLocation.GetTargetFile())
		}

		return err
	}

	logger.LogInfo("Compressed text file verified successfully: %s", destination.Import().Get().GetTargetFile())

	return nil
}
