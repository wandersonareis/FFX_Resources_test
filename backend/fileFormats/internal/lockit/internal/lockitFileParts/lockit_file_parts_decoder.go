package lockitFileParts

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
	"fmt"
)

type (
	ILockitFilePartsDecoder interface {
		DecodeFileParts() error
	}

	LockitFilePartsDecoder struct {
		Encoding  ffxencoding.IFFXTextLockitEncoding
		PartsList components.IList[LockitFileParts]
	}
)

func NewLockitFilePartsDecoder(partsList components.IList[LockitFileParts]) ILockitFilePartsDecoder {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	return &LockitFilePartsDecoder{
		Encoding:  encoding,
		PartsList: partsList,
	}
}

func (ld *LockitFilePartsDecoder) DecodeFileParts() error {
	//encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	//defer encoding.Dispose()

	if ld.PartsList.GetLength() == 0 {
		return fmt.Errorf("lockit file parts list is empty")
	}

	extractorFunc := func(index int, part LockitFileParts) {
		if index > 0 && index%2 == 0 {
			part.Extract(LocEnc, ld.Encoding)
		} else {
			part.Extract(FfxEnc, ld.Encoding)
		}
	}

	ld.PartsList.ParallelForEach(extractorFunc)

	return nil
}
