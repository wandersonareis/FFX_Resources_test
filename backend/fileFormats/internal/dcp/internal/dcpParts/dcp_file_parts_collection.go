package dcpParts

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/interfaces"
	"fmt"
)

func populateDcpPartsList(list components.IList[DcpFileParts], path, pattern string, formatter interfaces.ITextFormatter) error {
	err := components.PopulateFilePartsList(
		list,
		path,
		pattern,
		formatter,
		NewDcpFileParts)

	if err != nil {
		return fmt.Errorf(
			"error when finding dcp file parts: %s in path: %s",
			err.Error(),
			path)
	}

	return nil
}

func PopulateDcpBinaryFileParts(
	binaryPartsList components.IList[DcpFileParts],
	path string,
	formatter interfaces.ITextFormatter) error {
	return populateDcpPartsList(
		binaryPartsList,
		path,
		lib.DCP_FILE_BINARY_PARTS_PATTERN,
		formatter,
	)
}

func PopulateDcpTextFileParts(
	translatedPartsList components.IList[DcpFileParts],
	path string,
	formatter interfaces.ITextFormatter) error {
	return populateDcpPartsList(
		translatedPartsList,
		path,
		lib.DCP_FILE_TXT_PARTS_PATTERN,
		formatter,
	)
}
