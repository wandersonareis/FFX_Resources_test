package lockitFileEncoder

import ffxencoding "ffxresources/backend/core/encoding"

type (
	ILockitProcessingStrategy interface {
		Process(sourceFile, outputFile string, encoding ffxencoding.IFFXTextLockitEncoding) error
	}

	ILockitEncodingService interface {
		Process(sourceFile, outputFile string, encoding ffxencoding.IFFXTextLockitEncoding, strategy ILockitProcessingStrategy) error
	}

	LockitEncodingService struct {}
)

func NewLockitEncodingService() ILockitEncodingService {
	return &LockitEncodingService{}
}

func (svc *LockitEncodingService) Process(sourceFile, outputFile string, encoding ffxencoding.IFFXTextLockitEncoding, strategy ILockitProcessingStrategy) error {
	if err := strategy.Process(sourceFile, outputFile, encoding); err != nil {
		return err
	}
	return nil
}
