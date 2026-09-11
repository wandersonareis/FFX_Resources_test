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
	if gameVersion != 1 {
		common.LogVerbose("ReadNameOnlyLocalizations is only applicable for FFX (v1) game version.")
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		return NewNameOnlyTextObject(cBytes, sBytes, hLen, lang, gameVersion)
	}

	binaryDataFile := NewObjectBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
		gameVersion,
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
	if gameVersion != 1 && gameVersion != 2 {
		common.LogVerbose("ReadNameOnlyV2Localizations is only compatible with FFX (v1) or FFX-2 (v2) game versions.")
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		return NewNameOnlyTextObjectV2(cBytes, sBytes, hLen, lang, gameVersion)
	}

	binaryDataFile := NewObjectBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
		gameVersion,
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
	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		switch gameVersion {
		case 2:
			return NewCommandTextObjectV2(cBytes, sBytes, hLen, lang, gameVersion)
		case 1:
			return NewCommandTextObject(cBytes, sBytes, hLen, lang, interactions.NewInteractionService().FFXAppConfig().GetGameVersion())
		default:
			return nil, fmt.Errorf("CommandTextObject is only compatible with FFX (game version 1) or FFX-2 (game version 2), but got game version %d", gameVersion)
		}
	}

	binaryDataFile := NewObjectBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
		gameVersion,
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

func ReadJobLocalizations(patternPath string, effectSegmentPosition int64) datastore.IBinaryFile {
	gameVersion := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	if gameVersion != 2 {
		common.LogVerbose("ReadJobLocalizations is only compatible with FFX-2 (game version 2), but got game version %d", gameVersion)
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if gameVersion == 2 {
			return NewJobTextObject(cBytes, sBytes, hLen, effectSegmentPosition, lang, gameVersion)
		}
		return nil, fmt.Errorf("JobTextObject is only compatible with FFX-2 (game version 2), but got game version %d", gameVersion)
	}

	binaryDataFile := NewObjectBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
		gameVersion,
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

func ReadNameDescriptionEffectAbilitiesLocalizations(patternPath string, abilitiesCount int,
	effectSegmentPosition int64) datastore.IBinaryFile {
	gameVersion := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	if gameVersion != 2 {
		common.LogVerbose("ReadNameDescriptionEffectAbilitiesLocalizations is only compatible with FFX-2 (game version 2), but got game version %d", gameVersion)
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if gameVersion == 2 {
			return NewNameDescriptionEffectAbilityTextObject(cBytes, sBytes, hLen, abilitiesCount, effectSegmentPosition, lang, gameVersion)
		}
		return nil, fmt.Errorf("Name description effect abilities is only compatible with FFX-2 (game version 2), but got game version %d", gameVersion)
	}

	binaryDataFile := NewObjectBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
		gameVersion,
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

func ReadNameSensorScanLocalizations(patternPath string) datastore.IBinaryFile {
	gameVersion := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	if gameVersion != 1 {
		common.LogVerbose("ReadNameSensorScanLocalizations is only compatible with FFX (game version 1), but got game version %d", gameVersion)
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if gameVersion == 1 {
			return NewNameSensorScanTextObject(cBytes, sBytes, hLen, lang, gameVersion)
		}
		return nil, fmt.Errorf("Name sensor scan is only compatible with FFX (game version 1), but got game version %d", gameVersion)
	}

	binaryDataFile := NewObjectBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
		gameVersion,
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

func ReadWeaponNamesLocalizations(patternPath string) datastore.IBinaryFile {
	gameVersion := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	if gameVersion != 1 {
		common.LogVerbose("ReadWeaponNamesLocalizations is only compatible with FFX (game version 1), but got game version %d", gameVersion)
		return nil
	}
	
	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if gameVersion == 1 {
			return NewWeaponsNameTextObject(cBytes, sBytes, hLen, lang, gameVersion)
		}
		return nil, fmt.Errorf("Weapon names are only compatible with FFX (game version 1), but got game version %d", gameVersion)
	}

	binaryDataFile := NewObjectBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
		gameVersion,
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
