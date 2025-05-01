package lockit

import (
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/baseFormats"
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

	LockitFileCompressor struct {
		baseFormats.IBaseFileFormat

		lockitEncoding ffxencoding.IFFXTextLockitEncoding

		options core.ILockitFileOptions
		logger  logger.ILoggerHandler
	}
)

func NewLockitFileCompressor(
	source interfaces.ISource,
	destination locations.IDestination,
	lockitEncoding ffxencoding.IFFXTextLockitEncoding,
	fileOptions core.ILockitFileOptions,
	logger logger.ILoggerHandler,
) *LockitFileCompressor {
	return &LockitFileCompressor{
		IBaseFileFormat: baseFormats.NewFormatsBase(source, destination),
		lockitEncoding:  lockitEncoding,

		options: fileOptions,
		logger:  logger,
	}
}

func (lfc *LockitFileCompressor) Compress() error {
	partsLength := lfc.options.GetPartsLength()

	textTranslatedFilePartsList := components.NewList[lockitParts.LockitFileParts](partsLength)
	defer textTranslatedFilePartsList.Clear()

	if err := lfc.populateLockitTranslatedTextFileParts(textTranslatedFilePartsList); err != nil {
		return err
	}

	if err := lfc.ensureAllLockitTranslatedTextFileParts(textTranslatedFilePartsList, partsLength); err != nil {
		return err
	}

	binaryExtractedFilePartsList := components.NewList[lockitParts.LockitFileParts](partsLength)
	defer binaryExtractedFilePartsList.Clear()

	if err := lfc.populateLockitExtractedBinaryFilePartsList(binaryExtractedFilePartsList); err != nil {
		return err
	}

	if err := lfc.ensureAllLockitExtractedBinaryFileParts(binaryExtractedFilePartsList, partsLength); err != nil {
		return err
	}

	if err := lfc.encodingFilesParts(binaryExtractedFilePartsList); err != nil {
		return err
	}

	binaryTranslatedFilePartsList := components.NewList[lockitParts.LockitFileParts](partsLength)
	defer lfc.disposeList(binaryTranslatedFilePartsList)

	if err := lfc.populateLockitTranslatedBinaryFileParts(binaryTranslatedFilePartsList); err != nil {
		return err
	}

	if err := lfc.ensureAllLockitTranslatedBinaryFileParts(binaryTranslatedFilePartsList, partsLength); err != nil {
		return err
	}

	if err := lfc.joiningLockitBinaryFileParts(binaryTranslatedFilePartsList); err != nil {
		return err
	}

	return nil
}

func (lfc *LockitFileCompressor) populateLockitExtractedBinaryFilePartsList(extractedBinaryPartsList components.IList[lockitParts.LockitFileParts]) error {
	//extractedBinaryPartsList.Clear()

	lfc.logger.LogInfo("Populating lockit extracted binary file parts...")

	return lockitParts.PopulateLockitBinaryFileParts(
		extractedBinaryPartsList,
		lfc.GetDestination().Extract().GetTargetPath(),
	)
}

func (lfc *LockitFileCompressor) populateLockitTranslatedTextFileParts(translatedTextPartsList components.IList[lockitParts.LockitFileParts]) error {
	//translatedTextPartsList.Clear()

	lfc.logger.LogInfo("Populating lockit translated text file parts...")

	return lockitParts.PopulateLockitTextFileParts(
		translatedTextPartsList,
		lfc.GetDestination().Translate().GetTargetPath(),
	)
}

func (lfc *LockitFileCompressor) populateLockitTranslatedBinaryFileParts(translatedBinaryPartsList components.IList[lockitParts.LockitFileParts]) error {
	lfc.logger.LogInfo("Populating lockit translated binary file parts...")

	return lockitParts.PopulateLockitBinaryFileParts(
		translatedBinaryPartsList,
		lfc.GetDestination().Translate().GetTargetPath(),
	)
}

func (lfc *LockitFileCompressor) ensureAllLockitTranslatedTextFileParts(translatedTextPartsList components.IList[lockitParts.LockitFileParts], partsLength int) error {
	if translatedTextPartsList.GetLength() != partsLength {
		err := lib.ErrLockitFilePartsCountMismatch(partsLength, translatedTextPartsList.GetLength())

		lfc.logger.LogError(err, "error ensuring translated lockit text parts")

		return err
	}

	translatedTextList := components.NewList[string](partsLength)
	defer translatedTextList.Clear()

	translatedTextPartsList.ForEach(func(part lockitParts.LockitFileParts) {
		translatedTextList.Add(
			part.GetSource().Get().Path)
	})

	return lfc.validateLineBreaksCount(translatedTextList)
}

func (lfc *LockitFileCompressor) ensureAllLockitTranslatedBinaryFileParts(translatedBinaryPartsList components.IList[lockitParts.LockitFileParts], partsLength int) error {
	if translatedBinaryPartsList.GetLength() != partsLength {
		err := lib.ErrLockitFilePartsCountMismatch(partsLength, translatedBinaryPartsList.GetLength())

		lfc.logger.LogError(err, "error ensuring translated lockit binary parts")

		return err
	}

	translatedBinaryList := components.NewList[string](partsLength)
	defer translatedBinaryList.Clear()

	translatedBinaryPartsList.ForEach(func(part lockitParts.LockitFileParts) {
		translatedBinaryList.Add(
			part.GetSource().Get().Path)
	})

	return lfc.validateLineBreaksCount(translatedBinaryList)
}

func (l *LockitFileCompressor) ensureAllLockitExtractedBinaryFileParts(extractedBinaryPartsList components.IList[lockitParts.LockitFileParts], partsLength int) error {
	maxAttempts := 3
	for i := 0; i < maxAttempts; i++ {
		if extractedBinaryPartsList.GetLength() == partsLength {
			break
		}

		if err := l.extractMissingLockitBinaryFileParts(); err != nil {
			return err
		}

		extractedBinaryPartsList.Clear()

		if err := l.populateLockitExtractedBinaryFilePartsList(extractedBinaryPartsList); err != nil {
			return err
		}
	}

	if extractedBinaryPartsList.GetLength() != partsLength {
		err := lib.ErrLockitFilePartsCountMismatch(partsLength, extractedBinaryPartsList.GetLength())

		l.logger.LogError(err, "error ensuring extracted lockit binary parts")

		return err
	}

	extractedBinaryList := components.NewList[string](partsLength)
	defer extractedBinaryList.Clear()

	extractedBinaryPartsList.ForEach(func(part lockitParts.LockitFileParts) {
		extractedBinaryList.Add(
			part.GetSource().Get().Path)
	})

	if err := l.validateLineBreaksCount(extractedBinaryList); err != nil {
		return err
	}

	return nil
}

func (l *LockitFileCompressor) extractMissingLockitBinaryFileParts() error {
	l.logger.LogInfo("Missing lockit file parts detected. Attempting to extract...")

	splitter := internal.NewLockitFileSplitter()
	return splitter.FileSplitter(l.GetSource(), l.GetDestination().Extract(), l.options)
}

func (lfc *LockitFileCompressor) encodingFilesParts(extractedBinaryPartsList components.IList[lockitParts.LockitFileParts]) error {
	if extractedBinaryPartsList.IsEmpty() {
		return fmt.Errorf("error when encoding files parts: extracted binary parts list is empty")
	}

	lfc.logger.LogInfo("Encoding files parts to: %s", lfc.GetDestination().Import().GetTargetPath())

	filePartsEncoder := lockitParts.NewLockitFilePartsEncoder(lfc.logger)
	filePartsEncoder.EncodeFilesParts(extractedBinaryPartsList, lfc.lockitEncoding)

	return nil
}
func (lfc *LockitFileCompressor) joiningLockitBinaryFileParts(translatedBinaryPartsList components.IList[lockitParts.LockitFileParts]) error {
	filePartsJoiner := internal.NewLockitFileJoiner(lfc.logger)

	lfc.logger.LogInfo("Joining file parts inside file: %s", lfc.GetDestination().Import().GetTargetFile())

	if err := filePartsJoiner.JoinFileParts(lfc.GetDestination(), translatedBinaryPartsList, lfc.options); err != nil {
		return fmt.Errorf("error when joining lockit binary file parts: %s", err.Error())
	}

	return nil
}

func (lfc *LockitFileCompressor) validateLineBreaksCount(filesList components.IList[string]) error {
	filePartsIntegrity := integrity.NewLockitFilePartsIntegrity(lfc.logger)

	return filePartsIntegrity.ValidatePartsLineBreaksCount(
		filesList,
		lfc.options,
	)
}

func (lfc *LockitFileCompressor) disposeList(list components.IList[lockitParts.LockitFileParts]) {
	list.ForEach(func(part lockitParts.LockitFileParts) {
		os.Remove(part.GetSource().Get().Path)
	})

	list.Clear()
}
