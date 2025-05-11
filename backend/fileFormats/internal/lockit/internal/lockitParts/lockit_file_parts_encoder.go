package lockitParts

import (
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	lockitFileEncoder "ffxresources/backend/fileFormats/internal/lockit/internal/encoder"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"
)

type (
	ILockitFilePartsEncoder interface {
		EncodeFilesParts(binaryPartsList components.IList[LockitFileParts], encoding ffxencoding.IFFXTextLockitEncoding) error
	}
	LockitFilePartsEncoder struct {
		formatter interfaces.ITextFormatter
		log       logger.ILoggerHandler
	}
)

func NewLockitFilePartsEncoder(logger logger.ILoggerHandler) ILockitFilePartsEncoder {
	return &LockitFilePartsEncoder{
		formatter: formatters.NewTxtFormatter(),
		log:       logger,
	}
}

func (le *LockitFilePartsEncoder) EncodeFilesParts(partsList components.IList[LockitFileParts], lockitEncoding ffxencoding.IFFXTextLockitEncoding) error {
	if partsList.GetLength() == 0 {
		return fmt.Errorf("lockit file parts list is empty")
	}

	errChan := make(chan error, partsList.GetLength())

	utf8Enconding := lockitFileEncoder.NewLockitEncoderUTF8Strategy()
	ffxEnconding := lockitFileEncoder.NewLockitEncoderFFXStrategy()

	// Choosencodingstrategy returns the appropriate coding strategy
	// Based on the index: If the index is greater than zero and pair, the UTF-8 strategy returns;Otherwise, the FFX strategy returns.
	chooseEncodingStrategy := func(index int) lockitFileEncoder.ILockitProcessingStrategy {
		if index > 0 && index%2 == 0 {
			return utf8Enconding
		}
		return ffxEnconding
	}

	processLockitPartForEncoding := func(index int, part LockitFileParts) {
		encoderStrategy := chooseEncodingStrategy(index)
		errChan <- part.Compress(lockitEncoding, encoderStrategy)
	}

	partsList.ForIndex(processLockitPartForEncoding)

	close(errChan)

	for err := range errChan {
		if err != nil {
			le.log.LogError(err, "error when compressing lockit file parts: %s", err.Error())
		}
	}

	return nil
}
