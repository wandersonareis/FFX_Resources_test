package objectsfile

import (
	"fmt"

	"ffxresources/backend/common"
)

// Layouts por formato. Cada entrada mapeia versão -> campos (nome + gap).
// gap = bytes pulados DEPOIS do segmento (para posicional com skip).
var (
	CommandLayout = LayoutSet{
		common.GameVersionFFX: {
			{"name", 0},
			{"simplifiedName", 0},
			{"description", 0},
			{"simplifiedDescription", 0},
		},
	}

	CommandV2Layout = LayoutSet{
		common.GameVersionFFX2: {
			{"name", 0},
			{"description", 0},
		},
		common.GameVersionLastMiss: {
			{"name", 0},
			{"description", 0},
		},
	}

	NameOnlyLayout = LayoutSet{
		common.GameVersionFFX: {
			{"name", 0},
			{"simplifiedName", 0},
		},
	}

	NameOnlyV2Layout = LayoutSet{
		common.GameVersionFFX: {
			{"name", 0},
		},
		common.GameVersionFFX2: {
			{"name", 0},
		},
		common.GameVersionLastMiss: {
			{"name", 0},
		},
	}

	JobLayout = LayoutSet{
		common.GameVersionFFX2: {
			{"name", 0},
			{"description", 0},
			{"effect", 0},
		},
		common.GameVersionLastMiss: {
			{"name", 0},
			{"description", 0},
			{"effect", 0},
		},
	}

	MonsterLayout = LayoutSet{
		common.GameVersionFFX: {
			{"name", 0},
			{"sensorText", 0},
			{"simplifiedSensorText", 0},
			{"scanText", 0},
			{"simplifiedScanText", 0},
		},
	}

	LastMissionLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"description", 0},
			{"effect", 0},
			{"effectDescription", 0},
		},
	}

	// LastMissionMesLayout: name + description contíguos, sem skip.
	LastMissionMesLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"description", 0},
		},
	}

	LastMissionDressLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"description", 0},
			{"effect", 0},
		},
	}

	LastMissionCommandLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"description", 0},
			{"effect", 0},
		},
	}

	// LastMissionMonmagicLayout: name/description com skip 4, effect sem skip.
	LastMissionMonmagicLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 4},
			{"description", 4},
			{"effect", 0},
		},
	}

	LastMissionMonsterLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"description", 0},
		},
	}

	LastMissionPlayerLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"description", 0},
		},
	}

	LastMissionTrapLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"description", 0},
			{"effect", 0},
		},
	}

	LastMissionWarehouseLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"description", 0},
		},
	}
)

// Formatters centralizados. São funções reutilizáveis de ToString,
// independentes do tipo concreto, usando apenas FieldString (nil-safe).
// O fallback do KeyedStringFile já é o join por " | " quando nenhum é definido.
var (
	commandLegacyFmt = func(f *KeyedStringFile, lang string) string {
		return fmt.Sprintf("%s %s - %s %s",
			FieldString(f, "name", lang),
			FieldString(f, "simplifiedName", lang),
			FieldString(f, "description", lang),
			FieldString(f, "simplifiedDescription", lang))
	}

	nameOnlyLegacyFmt = func(f *KeyedStringFile, lang string) string {
		return fmt.Sprintf("%s %s",
			FieldString(f, "name", lang),
			FieldString(f, "simplifiedName", lang))
	}

	threePartLegacyFmt = func(f *KeyedStringFile, lang string) string {
		return fmt.Sprintf("%s - %s - %s",
			FieldString(f, "name", lang),
			FieldString(f, "description", lang),
			FieldString(f, "effect", lang))
	}

	// Formato único do NameDescriptionEffectAbilityTextObject (name - description).
	nameDescriptionFmt = func(f *KeyedStringFile, lang string) string {
		return formatNameDescription(
			FieldString(f, "name", lang),
			FieldString(f, "description", lang))
	}
)

// formatNameDescription é compartilhado entre o KeyedStringFile e o tipo
// concreto de abilities, centralizando o formato "name - description".
func formatNameDescription(name, description string) string {
	return fmt.Sprintf("%s - %s", name, description)
}

// weaponLegacyFmt espelha o ToString do WeaponsNameTextObject (formato único,
// só o weapon_refs do primeiro personagem). Mantido aqui para centralização.
func weaponLegacyFmt(w *WeaponsNameTextObject, lang string) string {
	return fmt.Sprintf("Weapons: %s", w.Names[Tidus].GetLocalizedString(lang))
}

// newKeyedString cria a instância aplicando um formatter legado opcional.
func newKeyedString(bytes, stringBytes []byte, headerLength int, languageCode string, version common.GameVersion, layouts LayoutSet, typeName string, formatter StringFormatter) (*KeyedStringFile, error) {
	f, err := NewKeyedStringFile(bytes, stringBytes, headerLength, languageCode, version, layouts, typeName)
	if err != nil {
		return nil, err
	}
	if formatter != nil {
		f.SetFormatter(formatter)
	}
	return f, nil
}
