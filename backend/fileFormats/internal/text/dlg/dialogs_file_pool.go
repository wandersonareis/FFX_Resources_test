package dlg

import (
	"ffxresources/backend/fileFormats/internal/text/lib/textVerifier"
	"sync"
)

var (
	extractorPool = &sync.Pool{
		New: func() interface{} {
			return NewDlgExtractor()
		},
	}
	compressorPool = &sync.Pool{
		New: func() interface{} {
			return newDlgCompressor()
		},
	}
	textVerifierPool = &sync.Pool{
		New: func() interface{} {
			return textVerifier.NewTextsVerify()
		},
	}
)

func rentDlgExtractor() IDlgExtractor {
	return extractorPool.Get().(IDlgExtractor)
}

func returnDlgExtractor(extractor IDlgExtractor) {
	extractorPool.Put(extractor)
}

func rentDlgCompressor() IDlgCompressor {
	return compressorPool.Get().(IDlgCompressor)
}

func returnDlgCompressor(compressor IDlgCompressor) {
	compressorPool.Put(compressor)
}

func rentTextVerifier() textVerifier.ITextVerifier {
	return textVerifierPool.Get().(textVerifier.ITextVerifier)
}

func returnTextVerifier(textVerifier textVerifier.ITextVerifier) {
	textVerifierPool.Put(textVerifier)
}