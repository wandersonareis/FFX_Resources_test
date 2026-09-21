package objectsfile

import (
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
)

// readWithLayout constrói um ObjectBinaryFile cujo creator instancia o
// KeyedStringFile genérico a partir do layout e formatter fornecidos.
func readWithLayout(patternPath string, gameVersion common.GameVersion, layouts LayoutSet, typeName string, formatter StringFormatter) datastore.IBinaryFile {
	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		return NewKeyedStringFile(cBytes, sBytes, hLen, lang, gameVersion, layouts, typeName)
	}

	return newObjectBinaryFile(patternPath, creatorFunc, gameVersion, formatter)
}

// newObjectBinaryFile monta, carrega e registra os objetos em datastore.Commands.
func newObjectBinaryFile(patternPath string, creatorFunc CreatorFunc, gameVersion common.GameVersion, formatter StringFormatter) datastore.IBinaryFile {
	if err := ffxencoding.EnsureAllCharsetsLoaded(gameVersion); err != nil {
		common.LogError("charset maps not loaded: %v", err)
		return nil
	}
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
func ReadNameOnlyLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, NameOnlyLayout, "NameOnlyTextObject", nameOnlyLegacyFmt)
}

// ReadNameOnlyV2Localizations lê arquivos com apenas um campo Name.
func ReadNameOnlyV2Localizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, NameOnlyV2Layout, "NameOnlyTextObjectV2", nil)
}

// ReadCommandLocalizations lê arquivos de comando (CommandTextObject v1/v2).
func ReadCommandLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	switch gameVersion {
	case common.GameVersionFFX:
		return readWithLayout(patternPath, gameVersion, CommandLayout, "CommandTextObject", commandLegacyFmt)
	case common.GameVersionFFX2, common.GameVersionLastMiss:
		return readWithLayout(patternPath, gameVersion, CommandV2Layout, "CommandTextObjectV2", nil)
	default:
		common.LogVerbose("CommandTextObject has unsupported game version %s", gameVersion)
		return nil
	}
}

// ReadLastMissionLocalizations lê lastmiss com Name/Description/Effect/EffectDescription.
func ReadLastMissionLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, LastMissionLayout, "LastMissionTextObject", nil)
}

// ReadLastMissionMesLocalizations lê lm_mes.bin: name + description contíguos, sem skip.
func ReadLastMissionMesLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, LastMissionMesLayout, "LastMissionMesTextObject", nil)
}

// ReadLastMissionCommandLocalizations lê lastmiss com skip posicional.
func ReadLastMissionCommandLocalizations(patternPath string, gameVersion common.GameVersion, skip int) datastore.IBinaryFile {
	layout := skipLayout(skip)
	return readWithLayout(patternPath, gameVersion, layout, "LastMissionCommand", threePartLegacyFmt)
}

// ReadLastMissionMonmagicLocalizations lê lm_monmagic.bin: name/description skip 4, effect sem skip.
func ReadLastMissionMonmagicLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, LastMissionMonmagicLayout, "LastMissionMonmagicTextObject", threePartLegacyFmt)
}

// ReadLastMissionMonsterLocalizations lê lm_monster.bin: name + description, sem skip.
func ReadLastMissionMonsterLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, LastMissionMonsterLayout, "LastMissionMonsterTextObject", nil)
}

// ReadLastMissionPlayerLocalizations lê lm_player.bin: name + description, sem skip.
func ReadLastMissionPlayerLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, LastMissionPlayerLayout, "LastMissionPlayerTextObject", nil)
}

// ReadLastMissionTrapLocalizations lê lm_trap.bin: name + description + effect, sem skip.
func ReadLastMissionTrapLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, LastMissionTrapLayout, "LastMissionTrapTextObject", threePartLegacyFmt)
}

// ReadLastMissionWarehouseLocalizations lê lm_warehouse.bin: name + description, sem skip.
func ReadLastMissionWarehouseLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, LastMissionWarehouseLayout, "LastMissionWarehouseTextObject", nil)
}

// ReadLastMissionDressLocalizations lê lastmiss dress sem skip.
func ReadLastMissionDressLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, LastMissionDressLayout, "LastMissionDress", threePartLegacyFmt)
}

// ReadAcessoryLocalizations lê acessórios (ffx2) com Effect em posição absoluta.
func ReadAccessoryLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, AccessoryLayout, "AccessoryTextObject", threePartLegacyFmt)
}

// ReadJobLocalizations lê job (ffx2) com Effect em posição absoluta.
func ReadJobLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, JobLayout, "JobTextObject", threePartLegacyFmt)
}

// ReadPlateLocalizations lê plate (ffx2) com Effect.
func ReadPlateLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, PlateLayout, "PlateTextObject", threePartLegacyFmt)
}

// ReadNameDescriptionEffectAbilitiesLocalizations lê plate (ffx2) com Effect posicional.
func ReadNameDescriptionEffectAbilitiesLocalizations(patternPath string, gameVersion common.GameVersion, abilitiesCount int,
	effectSegmentPosition int64) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, JobLayoutAt(effectSegmentPosition), "NameDescriptionEffectAbilityTextObject", threePartLegacyFmt)
}

// ReadMonsterLocalizations lê monster (ffx) com sensor/scan.
func ReadMonsterLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	return readWithLayout(patternPath, gameVersion, MonsterLayout, "MonsterTextObject", nil)
}

// ReadWeaponNamesLocalizations mantém o tipo específico de armas.
func ReadWeaponNamesLocalizations(patternPath string, gameVersion common.GameVersion) datastore.IBinaryFile {
	if gameVersion != common.GameVersionFFX {
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
