package objectsfile

import (
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/interactions"
)

// readWithLayout constrói um ObjectBinaryFile cujo creator instancia o
// KeyedStringFile genérico a partir do layout e formatter fornecidos.
func readWithLayout(patternPath string, layouts LayoutSet, typeName string, formatter StringFormatter) datastore.IBinaryFile {
	gameVersion := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	if _, ok := layouts[gameVersion.Normalize()]; !ok {
		common.LogVerbose("%s is not compatible with game version %s", typeName, gameVersion)
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		return NewKeyedStringFile(cBytes, sBytes, hLen, lang, gameVersion, layouts, typeName)
	}

	return newObjectBinaryFile(patternPath, creatorFunc, gameVersion, formatter)
}

// newObjectBinaryFile monta, carrega e registra os objetos em datastore.Commands.
func newObjectBinaryFile(patternPath string, creatorFunc CreatorFunc, gameVersion common.GameVersion, formatter StringFormatter) datastore.IBinaryFile {
	binaryDataFile := NewObjectBinaryFile(
		patternPath,
		func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
			obj, err := creatorFunc(cBytes, sBytes, hLen, lang)
			if err != nil {
				return nil, err
			}
			if formatter != nil {
				if kf, ok := obj.(*KeyedStringFile); ok {
					kf.SetFormatter(formatter)
				}
			}
			return obj, nil
		},
		common.DefaultLocalization,
		gameVersion,
	)

	if err := binaryDataFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading binary data: %v", err)
		return nil
	}

	if objects := binaryDataFile.GetObjects(); objects != nil && !objects.IsEmpty() {
		common.LogVerbose("Loaded %d objects with all localizations", objects.Len())
		datastore.Commands = objects
	}
	return binaryDataFile
}

// ReadNameOnlyLocalizations lê arquivos ffx com Name + SimplifiedName contíguos.
func ReadNameOnlyLocalizations(patternPath string) datastore.IBinaryFile {
	return readWithLayout(patternPath, NameOnlyLayout, "NameOnlyTextObject", nameOnlyLegacyFmt)
}

// ReadNameOnlyV2Localizations lê arquivos com apenas um campo Name.
func ReadNameOnlyV2Localizations(patternPath string) datastore.IBinaryFile {
	return readWithLayout(patternPath, NameOnlyV2Layout, "NameOnlyTextObjectV2", nil)
}

// ReadCommandLocalizations lê arquivos de comando (CommandTextObject v1/v2).
func ReadCommandLocalizations(patternPath string) datastore.IBinaryFile {
	gameVersion := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	switch gameVersion.Normalize() {
	case common.GameVersionFFX:
		return readWithLayout(patternPath, CommandLayout, "CommandTextObject", commandLegacyFmt)
	case common.GameVersionFFX2, common.GameVersionLastMiss:
		return readWithLayout(patternPath, CommandV2Layout, "CommandTextObjectV2", nil)
	default:
		common.LogVerbose("CommandTextObject has unsupported game version %s", gameVersion)
		return nil
	}
}

// ReadLastMissionLocalizations lê lastmiss com Name/Description/Effect/EffectDescription.
func ReadLastMissionLocalizations(patternPath string) datastore.IBinaryFile {
	return readWithLayout(patternPath, LastMissionLayout, "LastMissionTextObject", nil)
}

// ReadLastMissionMesLocalizations lê lm_mes.bin: name + description contíguos, sem skip.
func ReadLastMissionMesLocalizations(patternPath string) datastore.IBinaryFile {
	return readWithLayout(patternPath, LastMissionMesLayout, "LastMissionMesTextObject", nil)
}

// ReadLastMissionCommandLocalizations lê lastmiss com skip posicional.
func ReadLastMissionCommandLocalizations(patternPath string, skip int) datastore.IBinaryFile {
	layout := skipLayout(skip)
	return readWithLayout(patternPath, layout, "LastMissionCommand", threePartLegacyFmt)
}

// ReadLastMissionMonmagicLocalizations lê lm_monmagic.bin: name/description skip 4, effect sem skip.
func ReadLastMissionMonmagicLocalizations(patternPath string) datastore.IBinaryFile {
	return readWithLayout(patternPath, LastMissionMonmagicLayout, "LastMissionMonmagicTextObject", threePartLegacyFmt)
}

// ReadLastMissionMonsterLocalizations lê lm_monster.bin: name + description, sem skip.
func ReadLastMissionMonsterLocalizations(patternPath string) datastore.IBinaryFile {
	return readWithLayout(patternPath, LastMissionMonsterLayout, "LastMissionMonsterTextObject", nil)
}

// ReadLastMissionPlayerLocalizations lê lm_player.bin: name + description, sem skip.
func ReadLastMissionPlayerLocalizations(patternPath string) datastore.IBinaryFile {
	return readWithLayout(patternPath, LastMissionPlayerLayout, "LastMissionPlayerTextObject", nil)
}

// ReadLastMissionTrapLocalizations lê lm_trap.bin: name + description + effect, sem skip.
func ReadLastMissionTrapLocalizations(patternPath string) datastore.IBinaryFile {
	return readWithLayout(patternPath, LastMissionTrapLayout, "LastMissionTrapTextObject", threePartLegacyFmt)
}

// ReadLastMissionWarehouseLocalizations lê lm_warehouse.bin: name + description, sem skip.
func ReadLastMissionWarehouseLocalizations(patternPath string) datastore.IBinaryFile {
	return readWithLayout(patternPath, LastMissionWarehouseLayout, "LastMissionWarehouseTextObject", nil)
}

// ReadLastMissionDressLocalizations lê lastmiss dress com skip posicional.
func ReadLastMissionDressLocalizations(patternPath string, skip int) datastore.IBinaryFile {
	layout := skipLayout(skip)
	return readWithLayout(patternPath, layout, "LastMissionDress", threePartLegacyFmt)
}

// ReadJobLocalizations lê job (ffx2) com Effect em posição absoluta.
func ReadJobLocalizations(patternPath string, effectSegmentPosition int64) datastore.IBinaryFile {
	return readWithLayout(patternPath, JobLayoutAt(effectSegmentPosition), "JobTextObject", threePartLegacyFmt)
}

// ReadNameDescriptionEffectAbilitiesLocalizations lê plate (ffx2) com Effect posicional.
func ReadNameDescriptionEffectAbilitiesLocalizations(patternPath string, abilitiesCount int,
	effectSegmentPosition int64) datastore.IBinaryFile {
	return readWithLayout(patternPath, JobLayoutAt(effectSegmentPosition), "NameDescriptionEffectAbilityTextObject", threePartLegacyFmt)
}

// ReadMonsterLocalizations lê monster (ffx) com sensor/scan.
func ReadMonsterLocalizations(patternPath string) datastore.IBinaryFile {
	return readWithLayout(patternPath, MonsterLayout, "MonsterTextObject", nil)
}

// ReadWeaponNamesLocalizations mantém o tipo específico de armas.
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

// skipLayout monta campos na posição 0 com skip entre eles (como o segmentOffsets antigo).
func skipLayout(skip int) LayoutSet {
	return LayoutSet{
		common.GameVersionLastMiss: {
			{"name", skip},
			{"description", skip},
			{"effect", 0},
		},
	}
}

// JobLayoutAt monta Name/Description contíguos + Effect.
// TODO(lastmiss/job): o skip real do Effect ainda é posicional; medir a partir
// do fim do campo description (position 0 aqui) e atualizar quando o offset real
// do efect no layout intermediário for confirmado.
func JobLayoutAt(effectSegmentPosition int64) LayoutSet {
	gap := 0
	return LayoutSet{
		common.GameVersionFFX2: {
			{"name", 0},
			{"description", gap},
			{"effect", 0},
		},
		common.GameVersionLastMiss: {
			{"name", 0},
			{"description", gap},
			{"effect", 0},
		},
	}
}
