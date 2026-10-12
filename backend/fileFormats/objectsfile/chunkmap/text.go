package chunkmap

import (
	"fmt"

	"ffxresources/backend/common"
)

// TextField is a text value surfaced to callers. Index addresses the mapped
// text slot; Name is the Go struct field path for display and validation only.
type TextField struct {
	Index int
	Name  string
	Texts map[string]string
}

// TextExporter and TextImporter are intentionally separate capabilities.
type (
	TextExporter interface {
		ExportText(languageCode string) ([]TextField, error)
	}
	TextImporter interface {
		ImportText([]TextField) error
	}
)

// FileText contains the exported text fields for one chunk.
type FileText struct {
	Index  int
	Fields []TextField
}

func validateLanguage(languageCode string) error {
	if !common.IsSupportedLanguage(languageCode) {
		return fmt.Errorf("chunkmap: idioma nao suportado %q", languageCode)
	}
	return nil
}
