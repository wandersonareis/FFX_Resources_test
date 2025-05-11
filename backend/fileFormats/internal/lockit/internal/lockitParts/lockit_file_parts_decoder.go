package lockitParts

import (
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	lockitFileEncoder "ffxresources/backend/fileFormats/internal/lockit/internal/encoder"
	"fmt"
)

type (
	ILockitFilePartsDecoder interface {
		DecodeFileParts(partsList components.IList[LockitFileParts], lockitEncoding ffxencoding.IFFXTextLockitEncoding) error
	}

	LockitFilePartsDecoder struct{}
)

func NewLockitFilePartsDecoder() ILockitFilePartsDecoder {
	return &LockitFilePartsDecoder{}
}

func (ld *LockitFilePartsDecoder) DecodeFileParts(partsList components.IList[LockitFileParts], lockitEncoding ffxencoding.IFFXTextLockitEncoding) error {
	if partsList.GetLength() == 0 {
		return fmt.Errorf("lockit file parts list is empty")
	}

	errChan := make(chan error, partsList.GetLength())

	utf8Decoding := lockitFileEncoder.NewLockitDecoderUTF8Strategy()
	ffxDecoding := lockitFileEncoder.NewLockitDecoderFFXStrategy()

	// ChoosedeCodingStrategy returns the proper decoding strategy
	// Based on the index: If the index is greater than zero and pair, the UTF-8 strategy returns;Otherwise, the FFX strategy returns.
	chooseDecodingStrategy := func(index int) lockitFileEncoder.ILockitProcessingStrategy {
		if index > 0 && index%2 == 0 {
			return utf8Decoding
		}
		return ffxDecoding
	}

	processLockitPartForDecoding := func(index int, part LockitFileParts) {
		decoderStrategy := chooseDecodingStrategy(index)
		errChan <- part.Extract(lockitEncoding, decoderStrategy)
	}

	partsList.ForIndex(processLockitPartForDecoding)
	close(errChan)

	for err := range errChan {
		if err != nil {
			return fmt.Errorf("error when extracting lockit file parts: %w", err)
		}
	}

	return nil
}
