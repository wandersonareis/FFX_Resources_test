package objectsfile

import (
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/interactions"
	"fmt"
)

// ReadCommandLocalizations reads battle command data from the command.bin file
// and loads all available localizations for each command entry directly into COMMANDS.
//
// This function reads CommandDataObject entries containing name and description information
// for combat abilities and skills. Each command includes localized text for all supported
// languages in the game. The data is loaded directly into the COMMANDS variable.
//
// File format: name and description data
// Pattern path: ex: "battle/kernel/command.bin"
func ReadNameOnlyLocalizations(patternPath string) datastore.IBinaryFile {
	gameVersion := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	if common.ToInt(gameVersion) != 1 {
		common.LogVerbose("ReadNameOnlyLocalizations is only applicable for FFX (v1) game version.")
		return nil
	}

	version := common.ToInt(gameVersion)
	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		return NewNameOnlyTextObject(cBytes, sBytes, hLen, lang, version)
	}

	binaryDataFile := NewObjectBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
		version,
	)

	if err := binaryDataFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading commands binary data: %v", err)
		return nil
	}

	if objects := binaryDataFile.GetObjects(); objects != nil && !objects.IsEmpty() {
		common.LogVerbose("Loaded %d commands with all localizations", objects.Len())
		datastore.Commands = objects
	}
	return binaryDataFile
}

func ReadNameOnlyV2Localizations(patternPath string) datastore.IBinaryFile {
	gameVersion := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	version := common.ToInt(gameVersion)
	if version != 1 && version != 2 {
		common.LogVerbose("ReadNameOnlyV2Localizations is only compatible with FFX (v1) or FFX-2 (v2) game versions.")
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		return NewNameOnlyTextObjectV2(cBytes, sBytes, hLen, lang, version)
	}

	binaryDataFile := NewObjectBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
		version,
	)

	if err := binaryDataFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading commands binary data: %v", err)
		return nil
	}

	if objects := binaryDataFile.GetObjects(); objects != nil && !objects.IsEmpty() {
		common.LogVerbose("Loaded %d commands with all localizations", objects.Len())
		datastore.Commands = objects
	}
	return binaryDataFile
}

// ReadCommandLocalizations reads battle command data from the command.bin file
// and loads all available localizations for each command entry directly into COMMANDS.
//
// This function reads CommandDataObject entries containing name and description information
// for combat abilities and skills. Each command includes localized text for all supported
// languages in the game. The data is loaded directly into the COMMANDS variable.
//
// File format: name and description data
// Pattern path: ex: "battle/kernel/command.bin"
func ReadCommandLocalizations(patternPath string) datastore.IBinaryFile {
	gameVersion := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	version := common.ToInt(gameVersion)
	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		switch version {
		case 2:
			return NewCommandTextObjectV2(cBytes, sBytes, hLen, lang, version)
		case 1:
			return NewCommandTextObject(cBytes, sBytes, hLen, lang, version)
		default:
			return nil, fmt.Errorf("CommandTextObject is only compatible with FFX (game version 1) or FFX-2 (game version 2), but got game version %d", version)
		}
	}

	binaryDataFile := NewObjectBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
		version,
	)

	if err := binaryDataFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading commands binary data: %v", err)
		return nil
	}

	if objects := binaryDataFile.GetObjects(); objects != nil && !objects.IsEmpty() {
		common.LogVerbose("Loaded %d commands with all localizations", objects.Len())
		datastore.Commands = objects
	}
	return binaryDataFile
}
