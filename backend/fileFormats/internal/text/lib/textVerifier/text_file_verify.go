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

// Atribuindo funções a variáveis como se fossem enums
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

	//dv.log.LogInfo("Text file verified successfully: %s", extractLocation.GetTargetFile())

	return nil
}

func extractIntegrityCheck(source interfaces.ISource, destination locations.IDestination, segmentCounter ISegmentCounter, fileComparer IComparer, logger logger.ILoggerHandler) error {
	extractLocation := destination.Extract().Get()

	if err := segmentCounter.CountBinary(extractLocation.GetTargetFile()); err != nil {
		if err := os.Remove(extractLocation.GetTargetFile()); err != nil {
			logger.LogError(err, "failed to remove broken text file: %s", extractLocation.GetTargetFile())
			return fmt.Errorf("failed to verify text file: %s", extractLocation.GetTargetFile())
		}

		return err
	}

	if err := segmentCounter.CountText(extractLocation.GetTargetFile()); err != nil {
		if err := os.Remove(extractLocation.GetTargetFile()); err != nil {
			logger.LogError(err, "failed to remove broken text file: %s", extractLocation.GetTargetFile())
			return fmt.Errorf("failed to verify text file: %s", extractLocation.GetTargetFile())
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

	//createTemporaryFileInfo(source, destination)
	//defer extractLocation.DisposeTargetFile()

	if err := fileComparer.CompareTranslatedTextParts(translateLocation.GetTargetFile(), extractLocation.GetTargetFile()); err != nil {
		if err := os.Remove(importLocation.GetTargetFile()); err != nil {
			logger.LogError(err, "failed to remove broken text file: %s", importLocation.GetTargetFile())
			return fmt.Errorf("failed to verify text file: %s", importLocation.GetTargetFile())
		}

		return err
	}

	logger.LogInfo("Compressed text file verified successfully: %s", destination.Import().Get().GetTargetFile())

	return nil
}

/* func (dv *TextVerifier) VerifyCompress(source interfaces.ISource, destination locations.IDestination, extractor func(source interfaces.ISource, destination locations.IDestination) error) error {
	extractLocation := destination.Extract().Get()
	translateLocation := destination.Translate().Get()
	importLocation := destination.Import().Get()

	if err := importLocation.Validate(); err != nil {
		dv.log.LogInfo("Error on import location validation: %s", err)
		return err
	}

	dv.createTemporaryFileInfo(source, destination)
	defer extractLocation.DisposeTargetFile()

	if err := extractor(source, destination); err != nil {
		dv.log.LogError(err, "Error on reimported dialog file: %s", source.Get().Name)
		return err
	}

	if err := dv.filesComparer.CompareTranslatedTextParts(translateLocation.GetTargetFile(), extractLocation.GetTargetFile()); err != nil {
		dv.log.LogError(err, "Error on reimported text file: %s", source.Get().Name)
		return err
	}

	dv.log.LogInfo("Compressed text file verified successfully: %s", source.Get().Name)

	return nil
} */

/* func createTemporaryFileInfo(source interfaces.ISource, destination locations.IDestination) {
	tmp := common.NewTempProvider("tmp", ".txt")

	destination.Extract().Get().SetTargetFile(tmp.TempFile)
	destination.Extract().Get().SetTargetPath(tmp.TempFilePath)

	s := source.Get()
	s.Path = destination.Import().Get().GetTargetFile()
	source.Set(s)
} */
