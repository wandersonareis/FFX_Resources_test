package mt2

import (
	"ffxresources/backend/fileFormats/internal/text/textVerifier"
	"ffxresources/backend/logger"
	"sync"
)

type (
	krnlPool struct {
		pool *sync.Pool
		log  logger.ILoggerHandler
	}

	ExtractionServicePool struct {
		krnlPool
	}

	CompressionServicePool struct {
		krnlPool
	}

	TextVerificationServicePool struct {
		krnlPool
	}
)

func NewKrnlExtractorPool(logger logger.ILoggerHandler) *ExtractionServicePool {
	ep := &ExtractionServicePool{
		krnlPool{
			pool: &sync.Pool{},
			log:  logger},
	}

	ep.pool.New = func() any {
		return NewKrnlExtractor(logger)
	}
	return ep
}

func (ep *ExtractionServicePool) Rent() IKrnlExtractor {
	return ep.pool.Get().(IKrnlExtractor)
}

func (ep *ExtractionServicePool) Return(extractor IKrnlExtractor) {
	ep.pool.Put(extractor)
}

func NewKrnlCompressionServicePool(logger logger.ILoggerHandler) *CompressionServicePool {
	cp := &CompressionServicePool{
		krnlPool{
			pool: &sync.Pool{},
			log:  logger},
	}

	cp.pool.New = func() any {
		return NewKrnlCompressor(logger)
	}
	return cp
}

func (cp *CompressionServicePool) Rent() IKrnlCompressor {
	return cp.pool.Get().(IKrnlCompressor)
}

func (cp *CompressionServicePool) Return(compressor IKrnlCompressor) {
	cp.pool.Put(compressor)
}

func NewTextVerificationServicePool(logger logger.ILoggerHandler) *TextVerificationServicePool {
	tv := &TextVerificationServicePool{
		krnlPool{
			pool: &sync.Pool{},
			log:  logger},
	}

	tv.pool.New = func() any {
		return textVerifier.NewTextVerificationService(logger)
	}
	return tv
}

func (tv *TextVerificationServicePool) Rent() textVerifier.ITextVerificationService {
	return tv.pool.Get().(textVerifier.ITextVerificationService)
}

func (tv *TextVerificationServicePool) Return(textVerifier textVerifier.ITextVerificationService) {
	tv.pool.Put(textVerifier)
}

var (
	extractionServicePool       *ExtractionServicePool
	compressionServicePool      *CompressionServicePool
	textVerificationServicePool *TextVerificationServicePool
)

func InitExtractionServicePool(logger logger.ILoggerHandler) {
	extractionServicePool = NewKrnlExtractorPool(logger)
}

func InitCompressionServicePool(logger logger.ILoggerHandler) {
	compressionServicePool = NewKrnlCompressionServicePool(logger)
}

func InitTextVerificationServicePool(logger logger.ILoggerHandler) {
	textVerificationServicePool = NewTextVerificationServicePool(logger)
}

func RentKrnlExtractor() IKrnlExtractor {
	return extractionServicePool.Rent()
}

func ReturnKrnlExtractor(extractor IKrnlExtractor) {
	extractionServicePool.Return(extractor)
}

func RentKrnlCompressor() IKrnlCompressor {
	return compressionServicePool.Rent()
}

func ReturnKrnlCompressor(compressor IKrnlCompressor) {
	compressionServicePool.Return(compressor)
}

func RentTextVerifier() textVerifier.ITextVerificationService {
	return textVerificationServicePool.Rent()
}

func ReturnTextVerifier(textVerifier textVerifier.ITextVerificationService) {
	textVerificationServicePool.Return(textVerifier)
}
