package internal

import (
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/internal/text/encoding"
	"ffxresources/backend/interactions"
)

func KernelFileExtractor(fileInfo interactions.IGameDataInfo) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextKrnlEncoding()
	defer encoding.Dispose()

	extractLocation := fileInfo.GetExtractLocation()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return err
	}

	decoder := textsEncoding.NewDecoder()

	sourceFile := fileInfo.GetGameData().FullFilePath

	if err := decoder.KnrlDecoder(sourceFile, extractLocation.TargetFile, encoding); err != nil {
		return err
	}

	return nil
}
