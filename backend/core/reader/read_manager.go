package reader

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/macrodic"
	"fmt"
	"path/filepath"
)

func PrepareStringMacros(filename, localization string) error {
	resolvedFile, err := common.NewFileAccessor(filename)
	if err != nil {
		return fmt.Errorf("failed to resolve macro dictionary file: %w", err)
	}
	mdf := macrodic.NewMacroDictionaryFile(resolvedFile.ReadBytes(), localization)

	mdf.PublishStrings()
	return nil
}

func InitializeInternals() error {
	for _, cs := range common.Charsets {
		if err := PrepareCharset(cs); err != nil {
			return err
		}
	}

	for loc := range common.SupportedLanguages {
		path := filepath.Join(common.GetLocalizationRoot(loc), "menu", "macrodic.dcp")
		if err := PrepareStringMacros(path, loc); err != nil {
			return err
		}
	}
	/* for _, strings := range macrodic.MacroLookup {
		//fmt.Printf("MacroLookup[%d] s%dl%d\n", idx, idx/0x100, idx%0x100)
		for loc := range common.SupportedLanguages {
			content := strings.GetLocalizedContent(loc)
			if content == nil || content.IsEmpty() {
				continue
			}
			//macro := content.GetString()
			//fmt.Printf("  %s: %s\n", loc, macro)
		}
	} */

	return nil
}
