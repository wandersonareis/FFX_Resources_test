package ffxencoding

import (
	"ffxresources/backend/common"
	"ffxresources/backend/models"
	"fmt"
	"os"
	"strings"
)

type FFXTextEncodingFactory struct {
	FFXTextEncoding *ffxTextEncodingHelper
	EncodingFile    string
}

func NewFFXTextEncodingFactory() *FFXTextEncodingFactory {
	return &FFXTextEncodingFactory{
		FFXTextEncoding: newFFXTextEncodingHelper(),
	}
}

func (e *FFXTextEncodingFactory) FFXTextEncodingCodePage() {
	codePage := e.FFXTextEncoding.createFFXTextEncoding()

	tmpProvider := common.NewTempProvider()
	tmpProvider.ProvideTempFileWithExtension("ffx_text_encoding", ".tbs")

	e.EncodingFile = tmpProvider.File

	e.writeEncodingToFile(e.EncodingFile, codePage)
}

func (e *FFXTextEncodingFactory) CreateFFXTextDlgEncoding(dlgFileType models.NodeType) IFFXTextDlgEncoding {
	codePage := e.FFXTextEncoding.createFFXTextEncoding()

	tmpProvider := common.NewTempProvider()
	tmpProvider.ProvideTempFileWithExtension("ffx_text_encoding", ".tbs")

	e.writeEncodingToFile(tmpProvider.File, codePage)

	return newFFXTextDlgEncoding(tmpProvider.File, dlgFileType)
}

func (e *FFXTextEncodingFactory) CreateFFXTextKrnlEncoding() IFFXTextKrnlEncoding {
	codePage := e.FFXTextEncoding.createFFXTextEncoding()

	tmpProvider := common.NewTempProvider()
	tmpProvider.ProvideTempFileWithExtension("ffx_text_encoding", ".tbs")

	e.writeEncodingToFile(tmpProvider.File, codePage)

	return newFFXTextKrnlEncoding(tmpProvider.File)
}

func (e *FFXTextEncodingFactory) CreateFFXTextLocalizationEncoding() IFFXTextLockitEncoding {
	locCodePage := e.FFXTextEncoding.createFFXTextLocalizationEncoding()
	ffxCodePage := e.FFXTextEncoding.createFFXTextSimpleEncoding()

	locEncodingTemp := common.NewTempProvider().ProvideTempFileWithExtension("ffx_text_localization", ".tbs")
	ffxEcodingTemp := common.NewTempProvider().ProvideTempFileWithExtension("ffx_text_simple_encoding", ".tbs")

	e.writeEncodingToFile(locEncodingTemp.File, locCodePage)
	e.writeEncodingToFile(ffxEcodingTemp.File, ffxCodePage)

	return newFFXTextLockitEncoding(locEncodingTemp.File, ffxEcodingTemp.File)
}

func (e *FFXTextEncodingFactory) writeEncodingToFile(path string, codePage []string) {
	err := os.WriteFile(path, []byte(strings.Join(codePage, "\r\n")), 0644)
	if err != nil {
		fmt.Printf("failed to write to file: %v", err)
	}
}
