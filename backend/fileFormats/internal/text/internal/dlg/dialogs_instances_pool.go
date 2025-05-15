package dlg

import (
	"ffxresources/backend/fileFormats/internal/text/textverify"
	"ffxresources/backend/loggingService"
	"sync"
)

type (
	dlgPool struct {
		pool   *sync.Pool
		logger loggingService.ILoggerService
	}

	ExtractorPool struct {
		dlgPool
	}

	CompressorPool struct {
		dlgPool
	}

	TextVerifierPool struct {
		dlgPool
	}
)

func NewDlgExtractorPool(logger loggingService.ILoggerService) *ExtractorPool {
	ep := &ExtractorPool{
		dlgPool{
			pool:   &sync.Pool{},
			logger: logger},
	}

	ep.pool.New = func() interface{} {
		return NewDlgExtractor(logger)
	}
	return ep
}

func (ep *ExtractorPool) Rent() IDlgExtractor {
	return ep.pool.Get().(IDlgExtractor)
}

func (ep *ExtractorPool) Return(extractor IDlgExtractor) {
	ep.pool.Put(extractor)
}

func NewDlgCompressorPool(logger loggingService.ILoggerService) *CompressorPool {
	cp := &CompressorPool{
		dlgPool{
			pool:   &sync.Pool{},
			logger: logger},
	}

	cp.pool.New = func() interface{} {
		return NewDlgCompressor(logger)
	}
	return cp
}

func (cp *CompressorPool) Rent() IDlgCompressor {
	return cp.pool.Get().(IDlgCompressor)
}

func (cp *CompressorPool) Return(compressor IDlgCompressor) {
	cp.pool.Put(compressor)
}

func NewTextVerifierPool(logger loggingService.ILoggerService) *TextVerifierPool {
	tv := &TextVerifierPool{
		dlgPool{
			pool:   &sync.Pool{},
			logger: logger},
	}

	tv.pool.New = func() interface{} {
		return textverify.NewTextVerificationService()
	}
	return tv
}

func (tv *TextVerifierPool) Rent() textverify.ITextVerificationService {
	return tv.pool.Get().(textverify.ITextVerificationService)
}

func (tv *TextVerifierPool) Return(textVerifier textverify.ITextVerificationService) {
	tv.pool.Put(textVerifier)
}

var (
	extractorPool    *ExtractorPool
	compressorPool   *CompressorPool
	textVerifierPool *TextVerifierPool
)

func InitExtractorsPool(logger loggingService.ILoggerService) {
	extractorPool = NewDlgExtractorPool(logger)
}

func InitCompressorsPool(logger loggingService.ILoggerService) {
	compressorPool = NewDlgCompressorPool(logger)
}

func InitTextVerifierPool(logger loggingService.ILoggerService) {
	textVerifierPool = NewTextVerifierPool(logger)
}

func RentDlgExtractor() IDlgExtractor {
	return extractorPool.Rent()
}

func ReturnDlgExtractor(extractor IDlgExtractor) {
	extractorPool.Return(extractor)
}

func RentDlgCompressor() IDlgCompressor {
	return compressorPool.Rent()
}

func ReturnDlgCompressor(compressor IDlgCompressor) {
	compressorPool.Return(compressor)
}

func RentTextVerifier() textverify.ITextVerificationService {
	return textVerifierPool.Rent()
}

func ReturnTextVerifier(textVerifier textverify.ITextVerificationService) {
	textVerifierPool.Return(textVerifier)
}
