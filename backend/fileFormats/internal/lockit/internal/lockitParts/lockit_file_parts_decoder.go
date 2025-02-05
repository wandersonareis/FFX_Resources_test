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

	// extractorFunc applies different encodings based on the part’s index.
	// Even parts (greater than zero) are processed using UTF-8 encoding,
	// while index zero and odd parts use a custom FFX codepage.
	extractorFunc := func(index int, part LockitFileParts) {
		if index > 0 && index%2 == 0 {
			part.Extract(UTF8Encoding, lockitEncoding)
		} else {
			part.Extract(FFXEncoding, lockitEncoding)
		}
	}

	partsList.ParallelForEach(extractorFunc)

	return nil
}
