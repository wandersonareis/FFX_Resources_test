package models

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type spiraFilesTypes struct {
	regex     *regexp.Regexp
	spiratype NodeType
}

var spiraFiles = []spiraFilesTypes{
	{
		regex:     regexp.MustCompile(`^.*\.msb$`),
		spiratype: Tutorial,
	},
	{
		regex:     regexp.MustCompile(`macrodic\.dcp$`),
		spiratype: Dcp,
	},
	{
		regex:     regexp.MustCompile(`macrodic\.(00[0-9]|01[0-5])$`),
		spiratype: DcpParts,
	},
	{
		regex:     regexp.MustCompile(`macrodic\.(00[0-9]|01[0-5])\.txt$`),
		spiratype: DcpParts,
	},
	{
		regex:     regexp.MustCompile(`.*kernel.*\.bin$`),
		spiratype: Kernel,
	},
	{
		regex:     regexp.MustCompile(`.*loc_kit_ps3.*\.bin$`),
		spiratype: Lockit,
	},
	{
		regex:     regexp.MustCompile(`.*loc_kit_ps3.*\.part([0-9]{2})$`),
		spiratype: LockitParts,
	},
	{
		regex:     regexp.MustCompile(`.*ffx2.*(monlist|credits|crjiten|crcr0000)\.bin$`),
		spiratype: DialogsSpecial,
	},
}

func guessFileType(path string) NodeType {
	cleanPath := filepath.Clean(path)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return None
	}

	if info.IsDir() {
		return Folder
	}

	return guessSpiraFileType(cleanPath)
}

func guessSpiraFileType(path string) NodeType {
	lowerPath := strings.ToLower(path)

	for _, spiraFile := range spiraFiles {
		if spiraFile.regex.MatchString(lowerPath) {
			return spiraFile.spiratype
		}
	}

	if strings.HasSuffix(lowerPath, ".bin") {
		return Dialogs
	}

	return File
}
