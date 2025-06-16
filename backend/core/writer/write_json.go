package writer

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"os"
	"path/filepath"
)

type KeyItemData struct {
	ID          int               `json:"id"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
}

// Generic function to export data to JSON
func exportToJSON(data interface{}, fileName string, verboseMessage string) error {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := os.MkdirAll(editsPath, 0755); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	jsonPath := filepath.Join(editsPath, fileName)

	if err := common.SaveAsJSON(data, jsonPath); err != nil {
		return fmt.Errorf("error saving JSON data to file %s: %w", jsonPath, err)
	}

	if common.IsVerboseMode() {
		fmt.Printf("%s: %s\n", verboseMessage, jsonPath)
	}

	return nil
}

func WriteObjectsDataJSON(objects []*components.NameDescriptionTextObject, jsonFileName string) error {
	if common.IsVerboseMode() {
		fmt.Printf("Carregando objetos de dados com todas as localizações...\n")
	}

	if len(objects) == 0 {
		return fmt.Errorf("no objects loaded or empty")
	}

	var allKeyItemsData []KeyItemData

	for i, nameDescObj := range objects {
		if nameDescObj == nil {
			if common.IsVerboseMode() {
				fmt.Printf("Key item %d não possui NameDescriptionTextObject\n", i)
			}
			continue
		}
		keyItemData := KeyItemData{
			ID:          i,
			Name:        make(map[string]string),
			Description: make(map[string]string),
		}

		for locKey := range common.SupportedLanguages {
			if nameDescObj.Name != nil {
				nameText := nameDescObj.Name.GetLocalizedString(locKey)
				if nameText != "" {
					keyItemData.Name[locKey] = nameText
				}
			}

			if nameDescObj.Description != nil {
				descText := nameDescObj.Description.GetLocalizedString(locKey)
				if descText != "" {
					keyItemData.Description[locKey] = descText
				}
			}
		}

		if len(keyItemData.Name) > 0 || len(keyItemData.Description) > 0 {
			allKeyItemsData = append(allKeyItemsData, keyItemData)
		}
	}

	return exportToJSON(allKeyItemsData, jsonFileName, "Arquivo JSON de key items salvo em")
}

func WriteKeyItemsJSON() error {
	if common.IsVerboseMode() {
		fmt.Printf("Carregando key items com todas as localizações...\n")
	}

	if len(components.KEY_ITEMS) == 0 {
		return fmt.Errorf("key items not loaded or empty")
	}

	allNameDescObjects := make([]*components.NameDescriptionTextObject, 0, len(components.KEY_ITEMS))
	for _, keyItem := range components.KEY_ITEMS {
		if keyItem != nil && keyItem.NameDescriptionTextObject != nil {
			allNameDescObjects = append(allNameDescObjects, keyItem.NameDescriptionTextObject)
		}
	}

	err := WriteObjectsDataJSON(
		allNameDescObjects,
		"key_items_all_localizations.json",
	)
	if err != nil {
		return fmt.Errorf("error writing key items JSON: %w", err)
	}

	return nil
}

func WriteCommandJSON() error {
	if common.IsVerboseMode() {
		fmt.Printf("Carregando comandos com todas as localizações...\n")
	}

	if len(components.COMMANDS) == 0 {
		return fmt.Errorf("commands not loaded or empty")
	}

	allNameDescObjects := make([]*components.NameDescriptionTextObject, 0, len(components.COMMANDS))
	for _, command := range components.COMMANDS {
		if command != nil && command.NameDescriptionTextObject != nil {
			allNameDescObjects = append(allNameDescObjects, command.NameDescriptionTextObject)
		}
	}

	err := WriteObjectsDataJSON(
		allNameDescObjects,
		"commands_all_localizations.json",
	)
	if err != nil {
		return fmt.Errorf("error writing commands JSON: %w", err)
	}

	return nil
}
