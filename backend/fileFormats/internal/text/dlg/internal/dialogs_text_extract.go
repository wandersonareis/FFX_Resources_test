package internal

import (
	"ffxresources/backend/core/encoding"
	textsEncoding "ffxresources/backend/fileFormats/internal/text/encoding"
	"ffxresources/backend/interactions"
)

func DialogsFileExtractor(dialogsFileInfo interactions.IGameDataInfo) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(dialogsFileInfo.GetGameData().Type)
	defer encoding.Dispose()

	sourceFile := dialogsFileInfo.GetGameData().FullFilePath

	extractLocation := dialogsFileInfo.GetExtractLocation()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return err
	}

	decoder := textsEncoding.NewDecoder()

	if err := decoder.DlgDecoder(sourceFile, extractLocation.TargetFile, encoding); err != nil {
		return err
	}

	return nil
}
