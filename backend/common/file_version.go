package common

import (
	"path/filepath"
	"strings"
)

// FileVersionSuffix returns the file-name suffix for the active game version:
// "v2" for FFX-2, "v1" for FFX.
func FileVersionSuffix() string {
	if GetGameVersionString() == "ffx2" {
		return "v2"
	}
	return "v1"
}

// WithVersionSuffix inserts the active game-version suffix (v1/v2) into a filename,
// just before its extension, so exported/imported artifacts are unambiguously tied
// to the game they were extracted from.
func WithVersionSuffix(fileName string) string {
	suffix := FileVersionSuffix()
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	return base + "_" + suffix + ext
}
