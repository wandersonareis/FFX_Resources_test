package lockitParts

import (
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	lockitFileEncoder "ffxresources/backend/fileFormats/internal/lockit/internal/encoder"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/loggingService"
	"ffxresources/backend/models"
	"fmt"
)

type (
	ILockitFilePartsEncoder interface {
		EncodeFilesParts(partsList components.IList[LockitFileParts], lockitEncoding ffxencoding.IFFXTextLockitEncoding, gameVersion models.GameVersion) error
	}
	LockitFilePartsEncoder struct {
		formatter interfaces.ITextFormatter
		log       loggingService.ILoggerService
	}
)

func NewLockitFilePartsEncoder(logger loggingService.ILoggerService) ILockitFilePartsEncoder {
	return &LockitFilePartsEncoder{
		formatter: formatters.NewTxtFormatter(),
		log:       logger,
	}
}

func (le *LockitFilePartsEncoder) EncodeFilesParts(partsList components.IList[LockitFileParts], lockitEncoding ffxencoding.IFFXTextLockitEncoding, gameVersion models.GameVersion) error {
	if partsList.GetLength() == 0 {
		return fmt.Errorf("lockit file parts list is empty")
	}

	errChan := make(chan error, partsList.GetLength())

	partsList.ForIndex(func(index int, part LockitFileParts) {
		errChan <- part.Compress(lockitEncoding, le.chooseStrategy(index, gameVersion))
	})

	close(errChan)

	for err := range errChan {
		if err != nil {
			le.log.LogError(err, "error when compressing lockit file parts: %s", err.Error())
		}
	}

	return nil
}

func (le *LockitFilePartsEncoder) chooseStrategy(index int, gameVersion models.GameVersion) lockitFileEncoder.ILockitProcessingStrategy {
	getStrategyV1 := func(index int) lockitFileEncoder.ILockitProcessingStrategy {
		if index > 0 && index%2 == 0 {
			return lockitFileEncoder.NewLockitEncoderUTF8Strategy()
		}
		return lockitFileEncoder.NewLockitEncoderFFXStrategy()
	}

	getStrategyV2 := func(index int) lockitFileEncoder.ILockitProcessingStrategy {
		if index > 0 && index%2 == 0 {
			return lockitFileEncoder.NewLockitEncoderUTF8Strategy()
		}
		return lockitFileEncoder.NewLockitEncoderFFX2Strategy()
	}

	switch gameVersion {
	case models.FFX:
		return getStrategyV1(index)
	case models.FFX2:
		return getStrategyV2(index)
	default:
		return getStrategyV1(index)
	}
}
