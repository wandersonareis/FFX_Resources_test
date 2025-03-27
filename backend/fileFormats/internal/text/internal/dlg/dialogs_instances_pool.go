package dlg

import (
	"ffxresources/backend/fileFormats/internal/text/textVerifier"
	"ffxresources/backend/logger"
	"sync"
)

type (
	dlgPool struct {
		pool   *sync.Pool
		logger logger.ILoggerHandler
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

func NewDlgExtractorPool(logger logger.ILoggerHandler) *ExtractorPool {
	ep := &ExtractorPool{
		dlgPool{logger: logger},
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

func NewDlgCompressorPool(logger logger.ILoggerHandler) *CompressorPool {
	cp := &CompressorPool{
		dlgPool{logger: logger},
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

func NewTextVerifierPool(logger logger.ILoggerHandler) *TextVerifierPool {
	tv := &TextVerifierPool{
		dlgPool{logger: logger},
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

func InitExtractorsPool(logger logger.ILoggerHandler) {
	extractorPool = NewDlgExtractorPool(logger)
}

func InitCompressorsPool(logger logger.ILoggerHandler) {
	compressorPool = NewDlgCompressorPool(logger)
}

func InitTextVerifierPool(logger logger.ILoggerHandler) {
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

func RentTextVerifier() textVerifier.ITextVerifier {
	return textVerifierPool.Rent()
}

func ReturnTextVerifier(textVerifier textVerifier.ITextVerifier) {
	textVerifierPool.Return(textVerifier)
}
