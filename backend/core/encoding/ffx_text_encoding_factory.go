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

	tmpProvider := common.NewTempProvider("ffx_text_encoding", "tbs")

	e.EncodingFile = tmpProvider.TempFile

	e.writeEncodingToFile(e.EncodingFile, codePage)
}

func (e *FFXTextEncodingFactory) CreateFFXTextDlgEncoding(dlgFileType models.NodeType) IFFXTextDlgEncoding {
	codePage := e.FFXTextEncoding.createFFXTextEncoding()

	tmpProvider := common.NewTempProvider("ffx_text_encoding", "tbs")

	e.writeEncodingToFile(tmpProvider.TempFile, codePage)

	return newFFXTextDlgEncoding(tmpProvider.TempFile, dlgFileType)
}

func (e *FFXTextEncodingFactory) CreateFFXTextKrnlEncoding() IFFXTextKrnlEncoding {
	codePage := e.FFXTextEncoding.createFFXTextEncoding()

	tmpProvider := common.NewTempProvider("ffx_text_encoding", "tbs")

	e.writeEncodingToFile(tmpProvider.TempFile, codePage)

	return newFFXTextKrnlEncoding(tmpProvider.TempFile)
}

func (e *FFXTextEncodingFactory) CreateFFXTextLocalizationEncoding() IFFXTextLockitEncoding {
	locCodePage := e.FFXTextEncoding.createFFXTextLocalizationEncoding()
	ffxCodePage := e.FFXTextEncoding.createFFXTextSimpleEncoding()

	locEncodingTemp := common.NewTempProvider("ffx_text_localization", "tbs")

	ffxEcodingTemp := common.NewTempProvider("ffx_text_simple_encoding", "tbs")

	e.writeEncodingToFile(locEncodingTemp.TempFile, locCodePage)
	e.writeEncodingToFile(ffxEcodingTemp.TempFile, ffxCodePage)

	return newFFXTextLockitEncoding(locEncodingTemp.TempFile, ffxEcodingTemp.TempFile)
}

func (e *FFXTextEncodingFactory) writeEncodingToFile(path string, codePage []string) {
	err := os.WriteFile(path, []byte(strings.Join(codePage, "\r\n")), 0644)
	if err != nil {
		fmt.Printf("failed to write to file: %v", err)
	}
}
