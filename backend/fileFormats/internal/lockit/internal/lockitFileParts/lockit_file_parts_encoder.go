package lockitFileParts

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
)

type (
	ILockitFilePartsEncoder interface {
		EncodeFilesParts()
	}
	LockitFilePartsEncoder struct{
		Encoding ffxencoding.IFFXTextLockitEncoding
		PartsList components.IList[LockitFileParts]
	}
)

func NewLockitFilePartsEncoder(partsList components.IList[LockitFileParts]) ILockitFilePartsEncoder {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	return &LockitFilePartsEncoder{
		Encoding: encoding,
		PartsList: partsList,
	}
}

func (le *LockitFilePartsEncoder) EncodeFilesParts() {
	//encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	//defer le.Encoding.Dispose()

	compressorFunc := func(index int, part LockitFileParts) {
		if index > 0 && index%2 == 0 {
			part.Compress(LocEnc, le.Encoding)
		} else {
			part.Compress(FfxEnc, le.Encoding)
		}
	}

	le.PartsList.ParallelForEach(compressorFunc)

	/* if err := lj.Destination().Translate().Get().ProvideTargetPath(); err != nil {
		return err
	} */
}