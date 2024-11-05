package core

import (
	"ffxresources/backend/models"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type spiraFilesTypes struct {
	regex     *regexp.Regexp
	spiratype models.NodeType
}

var spiraFiles = []spiraFilesTypes{
	{
		regex:     regexp.MustCompile(`^.*\.msb$`),
		spiratype: models.Tutorial,
	},
	{
		regex:     regexp.MustCompile(`.*macrodic.*\.dcp$`),
		spiratype: models.Dcp,
	},
	{
		regex:     regexp.MustCompile(`.*macrodic.*\.00[0-6]$`),
		spiratype: models.DcpParts,
	},
	{
		regex:     regexp.MustCompile(`.*kernel.*\.bin$`),
		spiratype: models.Kernel,
	},
	{
		regex:     regexp.MustCompile(`.*loc_kit_ps3.*\.bin$`),
		spiratype: models.Lockit,
	},
	{
		regex:     regexp.MustCompile(`.*loc_kit_ps3.*\.part([0-9]{2})$`),
		spiratype: models.LockitParts,
	},
}

func guessFileType(path string) models.NodeType {
	cleanPath := filepath.Clean(path)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return models.None
	}

	if info.IsDir() {
		return models.Folder
	}

	return guessSpiraFileType(cleanPath)
}

func guessSpiraFileType(path string) models.NodeType {
	lowerPath := strings.ToLower(path)

	for _, spiraFile := range spiraFiles {
		if spiraFile.regex.MatchString(lowerPath) {
			return spiraFile.spiratype
		}
	}

	if strings.HasSuffix(lowerPath, ".bin") {
		return models.Dialogs
	}

	return models.File
}
