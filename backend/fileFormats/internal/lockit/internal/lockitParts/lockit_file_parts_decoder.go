package lockitParts

import (
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	lockitFileEncoder "ffxresources/backend/fileFormats/internal/lockit/internal/encoder"
	"ffxresources/backend/models"
	"fmt"
)

type (
	ILockitFilePartsDecoder interface {
		DecodeFileParts(partsList components.IList[LockitFileParts], lockitEncoding ffxencoding.IFFXTextLockitEncoding, gameVersion models.GameVersion) error
	}

	LockitFilePartsDecoder struct{}
)

func NewLockitFilePartsDecoder() ILockitFilePartsDecoder {
	return &LockitFilePartsDecoder{}
}

func (ld *LockitFilePartsDecoder) DecodeFileParts(partsList components.IList[LockitFileParts], lockitEncoding ffxencoding.IFFXTextLockitEncoding, gameVersion models.GameVersion) error {
	if partsList.GetLength() == 0 {
		return fmt.Errorf("lockit file parts list is empty")
	}

	errChan := make(chan error, partsList.GetLength())

	partsList.ForIndex(func(index int, part LockitFileParts) {
		errChan <- part.Extract(lockitEncoding, ld.chooseDecodingStrategy(index, gameVersion))
	})

	close(errChan)

	for err := range errChan {
		if err != nil {
			return fmt.Errorf("error when extracting lockit file parts: %w", err)
		}
	}

	return nil
}

func (le *LockitFilePartsDecoder) chooseDecodingStrategy(index int, gameVersion models.GameVersion) lockitFileEncoder.ILockitProcessingStrategy {
	getStrategyV1 := func(index int) lockitFileEncoder.ILockitProcessingStrategy {
		if index > 0 && index%2 == 0 {
			return lockitFileEncoder.NewLockitDecoderUTF8Strategy()
		}
		return lockitFileEncoder.NewLockitDecoderFFXStrategy()
	}

	getStrategyV2 := func(index int) lockitFileEncoder.ILockitProcessingStrategy {
		if index > 0 && index%2 == 0 {
			return lockitFileEncoder.NewLockitDecoderUTF8Strategy()
		}
		return lockitFileEncoder.NewLockitDecoderFFX2Strategy()
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