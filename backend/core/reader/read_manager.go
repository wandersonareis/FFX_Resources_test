package reader

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"path/filepath"
)

func PrepareStringMacros(filename, localization string) error {
	resolvedFile, err := common.NewFileAccessor(filename)
	if err != nil {
		return fmt.Errorf("failed to resolve macro dictionary file: %w", err)
	}
	data := components.FileToBytes(resolvedFile)

	mdf := components.NewMacroDictionaryFile(data, localization)

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
	for _, strings := range components.MacroLookup {
		//fmt.Printf("MacroLookup[%d] s%dl%d\n", idx, idx/0x100, idx%0x100)
		for loc := range common.SupportedLanguages {
			content := strings.GetLocalizedContent(loc)
			if content == nil || content.IsEmpty() {
				continue
			}
			//macro := content.GetString()
			//fmt.Printf("  %s: %s\n", loc, macro)
		}
	}
	return nil
}

func ReadCommands() []*components.CommandDataObject {
	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), "battle", "kernel", "command.bin")
	creator := func(data []byte, stringBytes []byte, localization string) *components.CommandDataObject {
		return components.NewCommandDataObject(data, stringBytes, localization)
	}

	abilities := components.ReadDataArray(filePath, common.DefaultLocalization, creator)

	return abilities
}

func ReadKeyItems() []*components.KeyItemDataObject {
	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), "battle", "kernel", "important.bin")

	creator := func(data []byte, stringBytes []byte, localization string) *components.KeyItemDataObject {
		return components.NewKeyItemDataObject(data, stringBytes, localization)
	}

	items := components.ReadDataArray(filePath, common.DefaultLocalization, creator)

	if common.IsVerboseMode() {
		if len(items) > 0 {
			for i, item := range items {
				if item != nil {
					fmt.Printf("Key Item %04X: %s\n", i*0x1000, item.String())
				}
			}
		}
	}
	return items
}

func populateDataObjectLocalizations[T components.DataObjectWithLocalizations](path string, objects []T, creator components.DataObjectCreator[T]) {
	if len(objects) == 0 {
		return
	}

	for locKey := range common.SupportedLanguages {
		fullPath := filepath.Join(common.GetLocalizationRoot(locKey), path)

		localizations := components.ReadDataArray(fullPath, locKey, creator)
		if localizations != nil {
			maxLen := min(len(localizations), len(objects))
			for i := range maxLen {
				objects[i].SetLocalizations(localizations[i])
			}
		}
	}
}

func ReadCommandsWithAllLocalizations() []*components.CommandDataObject {
	commands := ReadCommands()

	creator := func(data []byte, stringBytes []byte, localization string) *components.CommandDataObject {
		return components.NewCommandDataObject(data, stringBytes, localization)
	}

	populateDataObjectLocalizations("battle/kernel/command.bin", commands, creator)

	if common.IsVerboseMode() {
		fmt.Printf("Carregados %d comandos com todas as localizações\n", len(commands))
	}

	return commands
}

func ReadKeyItemsWithAllLocalizations() []*components.KeyItemDataObject {
	keyItems := ReadKeyItems()

	creator := func(data []byte, stringBytes []byte, localization string) *components.KeyItemDataObject {
		return components.NewKeyItemDataObject(data, stringBytes, localization)
	}

	populateDataObjectLocalizations("battle/kernel/important.bin", keyItems, creator)

	if common.IsVerboseMode() {
		fmt.Printf("Carregados %d key items com todas as localizações\n", len(keyItems))
	}

	return keyItems
}
