package mt2

import (
	"ffxresources/backend/fileFormats/internal/text/textVerifier"
	"ffxresources/backend/logger"
	"sync"
)

type (
	krnlPool struct {
		pool   *sync.Pool
		logger logger.ILoggerHandler
	}

	ExtractorPool struct {
		krnlPool
	}

	CompressorPool struct {
		krnlPool
	}

	TextVerifierPool struct {
		krnlPool
	}
)

func NewKrnlExtractorPool(logger logger.ILoggerHandler) *ExtractorPool {
	ep := &ExtractorPool{
		krnlPool{logger: logger},
	}

	ep.pool.New = func() interface{} {
		return newKrnlExtractor(logger)
	}
	return ep
}

func (ep *ExtractorPool) Rent() IKrnlExtractor {
	return ep.pool.Get().(IKrnlExtractor)
}

func (ep *ExtractorPool) Return(extractor IKrnlExtractor) {
	ep.pool.Put(extractor)
}

func NewKrnlCompressorPool(logger logger.ILoggerHandler) *CompressorPool {
	cp := &CompressorPool{
		krnlPool{logger: logger},
	}
	cp.pool.New = func() interface{} {
		return newKrnlCompressor(logger)
	}
	return cp
}

func (cp *CompressorPool) Rent() IKrnlCompressor {
	return cp.pool.Get().(IKrnlCompressor)
}

func (cp *CompressorPool) Return(compressor IKrnlCompressor) {
	cp.pool.Put(compressor)
}

func NewTextVerifierPool(logger logger.ILoggerHandler) *TextVerifierPool {
	tv := &TextVerifierPool{
		krnlPool{logger: logger},
	}
	tv.pool.New = func() interface{} {
		return textVerifier.NewTextsVerify(logger)
	}
	return tv
}

func (tv *TextVerifierPool) Rent() textVerifier.ITextVerifier {
	return tv.pool.Get().(textVerifier.ITextVerifier)
}

func (tv *TextVerifierPool) Return(textVerifier textVerifier.ITextVerifier) {
	tv.pool.Put(textVerifier)
}

var (
	extractorPool    *ExtractorPool
	compressorPool   *CompressorPool
	textVerifierPool *TextVerifierPool
)

func InitExtractorsPools(logger logger.ILoggerHandler) {
	extractorPool = NewKrnlExtractorPool(logger)
}

func InitCompressorsPools(logger logger.ILoggerHandler) {
	compressorPool = NewKrnlCompressorPool(logger)
}

func InitTextVerifiersPools(logger logger.ILoggerHandler) {
	textVerifierPool = NewTextVerifierPool(logger)
}

func RentKrnlExtractor() IKrnlExtractor {
	return extractorPool.Rent()
}

func ReturnKrnlExtractor(extractor IKrnlExtractor) {
	extractorPool.Return(extractor)
}

func RentKrnlCompressor() IKrnlCompressor {
	return compressorPool.Rent()
}

func ReturnKrnlCompressor(compressor IKrnlCompressor) {
	compressorPool.Return(compressor)
}

func RentTextVerifier() textVerifier.ITextVerifier {
	return textVerifierPool.Rent()
}

func ReturnTextVerifier(textVerifier textVerifier.ITextVerifier) {
	textVerifierPool.Return(textVerifier)
}
