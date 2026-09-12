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
	if gameVersion.Normalize() != common.GameVersionFFX {
		common.LogVerbose("ReadNameOnlyLocalizations is only applicable for FFX (ffx) game version.")
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
	if !gameVersion.Normalize().IsValid() {
		common.LogVerbose("ReadNameOnlyV2Localizations has unsupported game version %s.", gameVersion)
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
		switch gameVersion.Normalize() {
		case common.GameVersionFFX2, common.GameVersionLastMiss:
			return NewCommandTextObjectV2(cBytes, sBytes, hLen, lang, gameVersion)
		case common.GameVersionFFX:
			return NewCommandTextObject(cBytes, sBytes, hLen, lang, gameVersion)
		default:
			return nil, fmt.Errorf("CommandTextObject has unsupported game version %s", gameVersion)
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
	if gameVersion.Normalize() != common.GameVersionFFX2 && gameVersion.Normalize() != common.GameVersionLastMiss {
		common.LogVerbose("ReadJobLocalizations is only compatible with FFX-2 (ffx2), but got game version %s", gameVersion)
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		return NewJobTextObject(cBytes, sBytes, hLen, effectSegmentPosition, lang, gameVersion)
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
	if gameVersion.Normalize() != common.GameVersionFFX2 && gameVersion.Normalize() != common.GameVersionLastMiss {
		common.LogVerbose("ReadNameDescriptionEffectAbilitiesLocalizations is only compatible with FFX-2 (ffx2), but got game version %s", gameVersion)
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		return NewNameDescriptionEffectAbilityTextObject(cBytes, sBytes, hLen, abilitiesCount, effectSegmentPosition, lang, gameVersion)
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

func ReadMonsterLocalizations(patternPath string) datastore.IBinaryFile {
	gameVersion := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	if gameVersion.Normalize() != common.GameVersionFFX {
		common.LogVerbose("ReadMonsterLocalizations is only compatible with FFX (ffx), but got game version %s", gameVersion)
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		return NewMonsterTextObject(cBytes, sBytes, hLen, lang, gameVersion)
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
	if gameVersion.Normalize() != common.GameVersionFFX {
		common.LogVerbose("ReadWeaponNamesLocalizations is only compatible with FFX (ffx), but got game version %s", gameVersion)
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		return NewWeaponsNameTextObject(cBytes, sBytes, hLen, lang, gameVersion)
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
