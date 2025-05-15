package mt2

import (
	"ffxresources/backend/fileFormats/internal/text/textverify"
	"ffxresources/backend/loggingService"
	"sync"
)

type (
	krnlPool struct {
		pool *sync.Pool
		log  loggingService.ILoggerService
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

func NewKrnlExtractorPool(logger loggingService.ILoggerService) *ExtractionServicePool {
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

func NewKrnlCompressionServicePool(logger loggingService.ILoggerService) *CompressionServicePool {
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

func NewTextVerificationServicePool(logger loggingService.ILoggerService) *TextVerificationServicePool {
	tv := &TextVerificationServicePool{
		krnlPool{
			pool: &sync.Pool{},
			log:  logger},
	}

	tv.pool.New = func() any {
		return textverify.NewTextVerificationService()
	}
	return tv
}

func (tv *TextVerificationServicePool) Rent() textverify.ITextVerificationService {
	return tv.pool.Get().(textverify.ITextVerificationService)
}

func (tv *TextVerificationServicePool) Return(textVerifier textverify.ITextVerificationService) {
	tv.pool.Put(textVerifier)
}

var (
	extractionServicePool       *ExtractionServicePool
	compressionServicePool      *CompressionServicePool
	textVerificationServicePool *TextVerificationServicePool
)

func InitExtractionServicePool(logger loggingService.ILoggerService) {
	extractionServicePool = NewKrnlExtractorPool(logger)
}

func InitCompressionServicePool(logger loggingService.ILoggerService) {
	compressionServicePool = NewKrnlCompressionServicePool(logger)
}

func InitTextVerificationServicePool(logger loggingService.ILoggerService) {
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

func RentTextVerifier() textverify.ITextVerificationService {
	return textVerificationServicePool.Rent()
}

func ReturnTextVerifier(textVerifier textverify.ITextVerificationService) {
	textVerificationServicePool.Return(textVerifier)
}
