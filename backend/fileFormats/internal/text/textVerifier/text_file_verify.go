package textVerifier

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"ffxresources/backend/models"
	"fmt"
	"os"
	"sync"
	"time"
)

type TextIntegrityVerifyFunc func(source interfaces.ISource, destination locations.IDestination, segmentCounter ISegmentCounter, fileComparer IComparer, logger logger.ILoggerHandler) error

var (
	ExtractIntegrityCheck  TextIntegrityVerifyFunc = extractIntegrityCheck
	CompressIntegrityCheck TextIntegrityVerifyFunc = compressIntegrityCheck
)

type (
	ITextVerifier interface {
		Verify(interfaces.ISource, locations.IDestination, TextIntegrityVerifyFunc) error
		CompareTextSegmentsCount(binaryFile, textFile string, binaryType models.NodeType) error
	}

	TextVerifier struct {
		fileSegmentCounter  ISegmentCounter
		fileContentComparer IComparer

		initializeSegmentCounter sync.Once
		initializeFilesComparer  sync.Once

		log logger.ILoggerHandler
	}
)

func NewTextsVerify(logger logger.ILoggerHandler) *TextVerifier {
	return &TextVerifier{
		fileContentComparer: newPartComparer(),

		log: logger,
	}
}

func (dv *TextVerifier) Verify(source interfaces.ISource, destination locations.IDestination, verify TextIntegrityVerifyFunc) error {
	extractLocation := destination.Extract()
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

func (dv *TextVerifier) CompareTextSegmentsCount(binaryFile, textFile string, binaryType models.NodeType) error {
	dv.initializeSegmentCounter.Do(func() {
		dv.fileSegmentCounter = new(segmentCounter)
	})

	return dv.fileSegmentCounter.CompareTextSegmentsCount(binaryFile, textFile, binaryType)
}

func extractIntegrityCheck(
	source interfaces.ISource,
	destination locations.IDestination,
	segmentCounter ISegmentCounter,
	fileComparer IComparer,
	logger logger.ILoggerHandler) error {
	extractedFile := destination.Extract().GetTargetFile()

	sourceFileType := source.Get().Type
	sourceFile := source.Get().Path

	if err := segmentCounter.CompareTextSegmentsCount(sourceFile, extractedFile, sourceFileType); err != nil {
		if err := os.Remove(extractedFile); err != nil {
			logger.LogError(err, "failed to remove broken text file: %s", extractedFile)
			return fmt.Errorf("failed to integrity text file: %s", extractedFile)
		}
		return err
	}

	return nil
}

func compressIntegrityCheck(source interfaces.ISource, destination locations.IDestination, segmentCounter ISegmentCounter, fileComparer IComparer, logger logger.ILoggerHandler) error {
	extractLocation := destination.Extract()
	translateLocation := destination.Translate()
	importLocation := destination.Import()

	if err := importLocation.Validate(); err != nil {
		return err
	}

	if err := fileComparer.CompareTextPartsContents(translateLocation.GetTargetFile(), extractLocation.GetTargetFile()); err != nil {
		if err := common.RemoveFileWithRetries(importLocation.GetTargetFile(), 5, time.Second); err != nil {
			logger.LogError(err, "failed to remove broken text file: %s", importLocation.GetTargetFile())
		}

		return err
	}

	logger.LogInfo("Compressed text file verified successfully: %s", destination.Import().GetTargetFile())

	return nil
}
