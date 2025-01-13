package mt2

import (
	"ffxresources/backend/fileFormats/internal/text/lib/textVerifier"
	"sync"
)

var (
	extractorPool = &sync.Pool{
		New: func() interface{} {
			return newKrnlExtractor()
		},
	}
	compressorPool = &sync.Pool{
		New: func() interface{} {
			return newKrnlCompressor()
		},
	}
	textVerifierPool = &sync.Pool{
		New: func() interface{} {
			return textVerifier.NewTextsVerify()
		},
	}
)

func rentKrnlExtractor() IKrnlExtractor {
	return extractorPool.Get().(IKrnlExtractor)
}

func returnKrnlExtractor(extractor IKrnlExtractor) {
	extractorPool.Put(extractor)
}

func rentKrnlCompressor() IKrnlCompressor {
	return compressorPool.Get().(IKrnlCompressor)
}

func returnKrnlCompressor(compressor IKrnlCompressor) {
	compressorPool.Put(compressor)
}

func rentTextVerifier() textVerifier.ITextVerifier {
	return textVerifierPool.Get().(textVerifier.ITextVerifier)
}

func returnTextVerifier(textVerifier textVerifier.ITextVerifier) {
	textVerifierPool.Put(textVerifier)
}