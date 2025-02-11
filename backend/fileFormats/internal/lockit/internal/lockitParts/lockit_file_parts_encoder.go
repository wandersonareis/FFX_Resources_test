package lockitParts

import (
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
)

type (
	ILockitFilePartsEncoder interface {
		EncodeFilesParts(binaryPartsList components.IList[LockitFileParts], encoding ffxencoding.IFFXTextLockitEncoding)
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

func (le *LockitFilePartsEncoder) EncodeFilesParts(
	binaryPartsList components.IList[LockitFileParts],
	lockitEncoding ffxencoding.IFFXTextLockitEncoding) {
	errChan := make(chan error, binaryPartsList.GetLength())

	compressorFunc := func(index int, part LockitFileParts) {
		if index > 0 && index%2 == 0 {
			part.Compress(UTF8Encoding, lockitEncoding, errChan)
		} else {
			part.Compress(FFXEncoding, lockitEncoding, errChan)
		}
	}

	binaryPartsList.ParallelForIndex(compressorFunc)

	close(errChan)

	for err := range errChan {
		if err != nil {
			le.log.LogError(err, "error when compressing lockit file parts: %s", err.Error())
		}
	}
}
