package lockitParts

import (
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
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

	// extractorFunc applies different encodings based on the part’s index.
	// Even parts (greater than zero) are processed using UTF-8 encoding,
	// while index zero and odd parts use a custom FFX codepage.
	extractorFunc := func(index int, part LockitFileParts) {
		if index > 0 && index%2 == 0 {
			errChan <- part.Extract(UTF8Encoding, lockitEncoding)
		} else {
			errChan <- part.Extract(FFXEncoding, lockitEncoding)
		}
	}

	partsList.ForIndex(extractorFunc)
	close(errChan)

	for err := range errChan {
		if err != nil {
			return fmt.Errorf("error when extracting lockit file parts: %w", err)
		}
	}

	return nil
}
