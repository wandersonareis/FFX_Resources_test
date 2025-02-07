package lockit

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal"
	"ffxresources/backend/fileFormats/internal/lockit/internal/integrity"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitParts"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
	"os"
)

type (
	ILockitFileCompressor interface {
		Compress() error
		Dispose()
	}

	lockitFileCompressor struct {
		*base.FormatsBase

		binaryExtractedFilePartsList  components.IList[lockitParts.LockitFileParts]
		binaryTranslatedFilePartsList components.IList[lockitParts.LockitFileParts]
		textTranslatedFilePartsList   components.IList[lockitParts.LockitFileParts]

		filePartsEncoder   lockitParts.ILockitFilePartsEncoder
		filePartsIntegrity integrity.ILockitFilePartsIntegrity
		filePartsJoiner    internal.ILockitPartsJoiner
		lockitEncoding     ffxencoding.IFFXTextLockitEncoding

		//formatter interfaces.ITextFormatter
		options core.ILockitFileOptions
		logger  logger.ILoggerHandler
	}
)

func newLockitFileCompressor(
	source interfaces.ISource,
	destination locations.IDestination,
	lockitEncoding ffxencoding.IFFXTextLockitEncoding,
	fileOptions core.ILockitFileOptions,
	logger logger.ILoggerHandler,
) *lockitFileCompressor {
	return &lockitFileCompressor{
		FormatsBase:      base.NewFormatsBase(source, destination),
		filePartsEncoder: lockitParts.NewLockitFilePartsEncoder(logger),
		lockitEncoding:   lockitEncoding,

		//formatter: formatters.NewTxtFormatter(),
		options: fileOptions,
		logger:  logger,
	}
}

func (lfc *lockitFileCompressor) Compress() error {
	partsLength := lfc.options.GetPartsLength()

	lfc.initializeFileCompressor(partsLength)

	if err := lfc.populateLockitTranslatedTextFileParts(); err != nil {
		return err
	}

	if err := lfc.ensureAllLockitTranslatedTextFileParts(partsLength); err != nil {
		return err
	}

	if err := lfc.populateLockitExtractedBinaryFilePartsList(); err != nil {
		return err
	}

	if err := lfc.ensureAllLockitExtractedBinaryFileParts(partsLength); err != nil {
		return err
	}

	lfc.encodingFilesParts()

	if err := lfc.populateLockitTranslatedBinaryFileParts(); err != nil {
		return err
	}

	if err := lfc.ensureAllLockitTranslatedBinaryFileParts(partsLength); err != nil {
		return err
	}

	if err := lfc.joiningLockitBinaryFileParts(); err != nil {
		return err
	}

	return nil
}

func (lfc *lockitFileCompressor) initializeFileCompressor(partsLength int) {
	lfc.binaryExtractedFilePartsList = components.NewList[lockitParts.LockitFileParts](partsLength)
	lfc.binaryTranslatedFilePartsList = components.NewList[lockitParts.LockitFileParts](partsLength)
	lfc.textTranslatedFilePartsList = components.NewList[lockitParts.LockitFileParts](partsLength)
}

func (lfc *lockitFileCompressor) populateLockitExtractedBinaryFilePartsList() error {
	lfc.binaryExtractedFilePartsList.Clear()

	lfc.logger.LogInfo("Populating lockit extracted binary file parts...")

	return lockitParts.PopulateLockitBinaryFileParts(
		lfc.binaryExtractedFilePartsList,
		lfc.Destination().Extract().Get().GetTargetPath(),
	)
}

func (lfc *lockitFileCompressor) populateLockitTranslatedTextFileParts() error {
	lfc.textTranslatedFilePartsList.Clear()

	lfc.logger.LogInfo("Populating lockit translated text file parts...")

	return lockitParts.PopulateLockitTextFileParts(
		lfc.textTranslatedFilePartsList,
		lfc.Destination().Translate().Get().GetTargetPath(),
	)
}

func (lfc *lockitFileCompressor) populateLockitTranslatedBinaryFileParts() error {
	lfc.binaryTranslatedFilePartsList.Clear()

	lfc.logger.LogInfo("Populating lockit translated binary file parts...")

	return lockitParts.PopulateLockitBinaryFileParts(
		lfc.binaryTranslatedFilePartsList,
		lfc.Destination().Translate().Get().GetTargetPath(),
	)
}

func (lfc *lockitFileCompressor) ensureAllLockitTranslatedTextFileParts(partsLength int) error {
	if lfc.textTranslatedFilePartsList.GetLength() != partsLength {
		err := lib.ErrLockitFilePartsCountMismatch(partsLength, lfc.textTranslatedFilePartsList.GetLength())

		lfc.logger.LogError(err, "error ensuring translated lockit text parts")

		return err
	}

	translatedTextList := components.NewList[string](partsLength)
	defer translatedTextList.Clear()

	lfc.textTranslatedFilePartsList.ForEach(func(part lockitParts.LockitFileParts) {
		//part.Destination().InitializeLocations(part.Source(), lfc.formatter)
		translatedTextList.Add(
			part.Source().Get().Path)
	})

	return lfc.validateLineBreaksCount(translatedTextList)
}

func (lfc *lockitFileCompressor) ensureAllLockitTranslatedBinaryFileParts(partsLength int) error {
	if lfc.binaryTranslatedFilePartsList.GetLength() != partsLength {
		err := lib.ErrLockitFilePartsCountMismatch(partsLength, lfc.binaryTranslatedFilePartsList.GetLength())

		lfc.logger.LogError(err, "error ensuring translated lockit binary parts")

		return err
	}

	translatedBinaryList := components.NewList[string](partsLength)
	defer translatedBinaryList.Clear()

	lfc.binaryTranslatedFilePartsList.ForEach(func(part lockitParts.LockitFileParts) {
		//part.Destination().InitializeLocations(part.Source(), lfc.formatter)
		translatedBinaryList.Add(
			part.Source().Get().Path)
	})

	return lfc.validateLineBreaksCount(translatedBinaryList)
}

func (l *lockitFileCompressor) ensureAllLockitExtractedBinaryFileParts(partsLength int) error {
	maxAttempts := 3
	for i := 0; i < maxAttempts; i++ {
		if l.binaryExtractedFilePartsList.GetLength() == partsLength {
			break
		}

		if err := l.extractMissingLockitBinaryFileParts(); err != nil {
			return err
		}

		if err := l.populateLockitExtractedBinaryFilePartsList(); err != nil {
			return err
		}
	}

	if l.binaryExtractedFilePartsList.GetLength() != partsLength {
		err := lib.ErrLockitFilePartsCountMismatch(partsLength, l.binaryExtractedFilePartsList.GetLength())

		l.logger.LogError(err, "error ensuring extracted lockit binary parts")

		return err
	}

	extractedBinaryList := components.NewList[string](partsLength)
	defer extractedBinaryList.Clear()

	l.binaryExtractedFilePartsList.ForEach(func(part lockitParts.LockitFileParts) {
		extractedBinaryList.Add(
			part.Source().Get().Path)
	})

	if err := l.validateLineBreaksCount(extractedBinaryList); err != nil {
		return err
	}

	return nil
}

func (l *lockitFileCompressor) extractMissingLockitBinaryFileParts() error {
	l.logger.LogInfo("Missing lockit file parts detected. Attempting to extract...")

	splitter := internal.NewLockitFileSplitter()
	return splitter.FileSplitter(l.Source(), l.Destination().Extract().Get(), l.options)
}

func (lfc *lockitFileCompressor) encodingFilesParts() {
	lfc.logger.LogInfo("Encoding files parts to: %s", lfc.Destination().Import().Get().GetTargetPath())

	lfc.filePartsEncoder.EncodeFilesParts(lfc.binaryExtractedFilePartsList, lfc.lockitEncoding)
}
func (lfc *lockitFileCompressor) joiningLockitBinaryFileParts() error {
	lfc.filePartsJoiner = internal.NewLockitFileJoiner(lfc.logger)
	defer func() {
		lfc.filePartsJoiner = nil
	}()

	lfc.logger.LogInfo("Joining file parts inside file: %s", lfc.Destination().Import().Get().GetTargetFile())

	if err := lfc.filePartsJoiner.JoinFileParts(lfc.Destination(), lfc.binaryTranslatedFilePartsList, lfc.options); err != nil {
		return fmt.Errorf("error when joining lockit binary file parts: %s", err.Error())
	}

	return nil
}

func (lfc *lockitFileCompressor) validateLineBreaksCount(filesList components.IList[string]) error {
	if lfc.filePartsIntegrity == nil {
		lfc.filePartsIntegrity = integrity.NewLockitFilePartsIntegrity(lfc.logger)
	}

	err := lfc.filePartsIntegrity.ValidatePartsLineBreaksCount(
		filesList,
		lfc.options,
	)

	return err
}

func (lfc *lockitFileCompressor) Dispose() {
	lfc.binaryTranslatedFilePartsList.ForEach(func(part lockitParts.LockitFileParts) {
		os.Remove(part.Source().Get().Path)
	})

	lfc.binaryExtractedFilePartsList.Clear()
	lfc.binaryTranslatedFilePartsList.Clear()
	lfc.textTranslatedFilePartsList.Clear()

	if lfc.filePartsIntegrity != nil {
		lfc.filePartsIntegrity = nil
	}

	if lfc.filePartsEncoder != nil {
		lfc.filePartsEncoder = nil
	}
}
