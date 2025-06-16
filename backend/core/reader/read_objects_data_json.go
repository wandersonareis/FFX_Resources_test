package reader

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"path/filepath"
)

type ObjectsData struct {
	ID          int               `json:"id"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
}

func extractNameDescObjects[T components.NameDescriptionGetter](items []T) []*components.NameDescriptionTextObject {
	var result []*components.NameDescriptionTextObject
	for _, item := range items {
		if obj := item.GetNameDescriptionTextObject(); obj != nil {
			result = append(result, obj)
		}
	}
	return result
}

/*
JSON KEY ITEMS EDITOR FUNCTIONS
===============================

This section contains JSON functions for key items editor.
These functions process the single JSON file created by WriteKeyItemsJSON.

1. ProcessKeyItemsJsonFile(print) - Processes the key_items_all_localizations.json file
  - Reads the specific JSON file created by WriteKeyItemsJSON
  - Processes all key items from the single JSON file
  - Applies changes back to the KEY_ITEMS component

2. editAndSaveKeyItemsFromJSON(print, path) - Processes a single JSON file
  - Reads JSON content and parses it into KeyItemData structure
  - Maps JSON data back to KeyItemDataObject
  - Updates localized content for each language

Usage:

	ProcessKeyItemsJsonFile(true)  // Process key_items_all_localizations.json with debug output
	ProcessKeyItemsJsonFile(false) // Process silently
*/
func ProcessKeyItemsJsonFile() error {
	jsonFilePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", "key_items_all_localizations.json")
	if !common.IsPathExists(jsonFilePath) {
		fmt.Printf("Arquivo JSON não encontrado: %s\n", jsonFilePath)
		return fmt.Errorf("JSON file not found: %s", jsonFilePath)
	}

	if common.IsVerboseMode() {
		fmt.Printf("Processando arquivo JSON de key items: %s\n", jsonFilePath)
	}

	keyItemsData, err := common.ReadJsonFile[[]ObjectsData](jsonFilePath)
	if err != nil {
		fmt.Printf("Error on reading JSON file %s: %v\n", jsonFilePath, err)
		return err
	}

	if common.IsVerboseMode() {
		fmt.Printf("Encontrados %d key items no arquivo JSON\n", len(keyItemsData))
	}

	nameDescObjects := extractNameDescObjects(components.KEY_ITEMS)

	if err := processAndSaveNameDescriptions(keyItemsData, nameDescObjects); err != nil {
		fmt.Printf("Erro ao processar arquivo JSON de key items: %v\n", err)
		return err
	}

	if common.IsVerboseMode() {
		fmt.Printf("Key items processados com sucesso!\n")
	}

	return nil
}

/*
JSON COMMANDS EDITOR FUNCTIONS
===============================

This section contains JSON functions for commands editor.
These functions process the single JSON file created by WriteCommandsJSON.

1. ProcessCommandJsonFile() - Processes the commands_all_localizations.json file
  - Reads the specific JSON file created by WriteCommandsJSON
  - Processes all commands from the single JSON file
  - Applies changes back to the COMMANDS component

2. editAndSaveNameDescriptionFromJSON(itemsData, objects) - Processes JSON data and updates objects
  - Reads JSON content and parses it into ObjectsData structure
  - Maps JSON data back to NameDescriptionTextObject
  - Updates localized content for each language

Usage:

	ProcessCommandJsonFile()  // Process commands_all_localizations.json with verbose output if enabled
*/
func ProcessCommandJsonFile() error {
	jsonFilePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", "commands_all_localizations.json")
	if !common.IsPathExists(jsonFilePath) {
		if common.IsVerboseMode() {
			fmt.Printf("Arquivo JSON não encontrado: %s\n", jsonFilePath)
		}
		return fmt.Errorf("JSON file not found: %s", jsonFilePath)
	}

	if common.IsVerboseMode() {
		fmt.Printf("Processando arquivo JSON de comandos: %s\n", jsonFilePath)
	}

	commandsData, err := common.ReadJsonFile[[]ObjectsData](jsonFilePath)
	if err != nil {
		if common.IsVerboseMode() {
			fmt.Printf("Erro ao ler arquivo JSON %s: %v\n", jsonFilePath, err)
		}
		return err
	}

	if common.IsVerboseMode() {
		fmt.Printf("Encontrados %d comandos no arquivo JSON\n", len(commandsData))
	}

	nameDescObjects := extractNameDescObjects(components.COMMANDS)

	if err := processAndSaveNameDescriptions(commandsData, nameDescObjects); err != nil {
		if common.IsVerboseMode() {
			fmt.Printf("Erro ao processar arquivo JSON de comandos: %v\n", err)
		}
		return err
	}

	if common.IsVerboseMode() {
		fmt.Printf("Comandos processados com sucesso!\n")
	}

	return nil
}

func createNewKeyedString(text string, charset string) *components.KeyedString {
	return &components.KeyedString{
		Charset: charset,
		Offset:  0,
		Key:     0,
		Bytes:   components.StringToBytes(text, charset),
	}
}

func updateOrCreateName(target *components.NameDescriptionTextObject, languageCode, text string) {
	existingContent := target.Name.GetLocalizedContent(languageCode)
	charset := common.LanguageCodeToCharset(languageCode)

	if existingContent != nil {
		existingContent.SetString(text, charset)
	} else {
		target.Name.SetLocalizedContent(languageCode, createNewKeyedString(text, charset))
	}

	if common.IsVerboseMode() {
		fmt.Printf("  Nome atualizado (%s): %s\n", languageCode, text)
	}
}

func updateOrCreateDescription(target *components.NameDescriptionTextObject, languageCode, text string) {
	existingContent := target.Description.GetLocalizedContent(languageCode)
	charset := common.LanguageCodeToCharset(languageCode)

	if existingContent != nil {
		existingContent.SetString(text, charset)
	} else {
		target.Description.SetLocalizedContent(languageCode, createNewKeyedString(text, charset))
	}

	if common.IsVerboseMode() {
		fmt.Printf("  Descrição atualizada (%s): %s\n", languageCode, text)
	}
}

func updateNameEntry(sourceData ObjectsData, targetObject *components.NameDescriptionTextObject) {
	if len(sourceData.Name) == 0 || targetObject.Name == nil {
		if common.IsVerboseMode() {
			fmt.Printf("Nenhum nome encontrado para o item %d, ignorando...\n", sourceData.ID)
		}
		return
	}

	for languageCode, nameText := range sourceData.Name {
		if nameText == "" {
			if common.IsVerboseMode() {
				fmt.Printf("  Nome vazio para linguagem %s, ignorando...\n", languageCode)
			}
			continue
		}

		if !common.IsSupportedLanguage(languageCode) {
			if common.IsVerboseMode() {
				fmt.Printf("  Localização não reconhecida para nome: %s\n", languageCode)
			}
			continue
		}

		updateOrCreateName(targetObject, languageCode, nameText)
	}
}

func updateDescriptionEntry(sourceData ObjectsData, targetObject *components.NameDescriptionTextObject) {
	if len(sourceData.Description) == 0 || targetObject.Description == nil {
		if common.IsVerboseMode() {
			fmt.Printf("  Nenhuma descrição encontrada para o item %d, ignorando...\n", sourceData.ID)
		}
		return
	}

	for languageCode, descriptionText := range sourceData.Description {
		if descriptionText == "" {
			if common.IsVerboseMode() {
				fmt.Printf("  Descrição vazia para linguagem %s, ignorando...\n", languageCode)
			}
			continue
		}

		if !common.IsSupportedLanguage(languageCode) {
			if common.IsVerboseMode() {
				fmt.Printf("  Localização não reconhecida para descrição: %s\n", languageCode)
			}
			continue
		}

		updateOrCreateDescription(targetObject, languageCode, descriptionText)
	}
}

func processAndSaveNameDescriptions(itemsData []ObjectsData, objects []*components.NameDescriptionTextObject) error {
	for _, jsonEntry := range itemsData {
		jsonEntryID := jsonEntry.ID

		if jsonEntryID < 0 || jsonEntryID >= len(objects) {
			if common.IsVerboseMode() {
				fmt.Printf("Key item ID fora do range: %d\n", jsonEntryID)
			}
			continue
		}

		nameDescObj := objects[jsonEntryID]
		if nameDescObj == nil {
			if common.IsVerboseMode() {
				fmt.Printf("Key item não encontrado: %d\n", jsonEntryID)
			}
			continue
		}

		if common.IsVerboseMode() {
			fmt.Printf("Processando key item %d\n", jsonEntryID)
		}

		updateNameEntry(jsonEntry, nameDescObj)
		updateDescriptionEntry(jsonEntry, nameDescObj)
	}
	return nil
}
