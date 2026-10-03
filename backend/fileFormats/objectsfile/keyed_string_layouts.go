package objectsfile

import (
	"fmt"

	"ffxresources/backend/common"
)

// Layouts por formato. Cada entrada mapeia versão -> campos (nome + gap).
// gap = bytes pulados DEPOIS do segmento (para posicional com skip).
//
// Os nomes seguem o fluxo novo (snake_case dos campos do struct C# —
// ver "Convenções" nos pacotes schema/ffx e schema/ffx2): pares
// standard/simplified ganham sufixo `_simplified`; campos aninhados levam
// o prefixo do pai (`creature_data_help`).
var (
	// CommandLayout: pares name + desc (C# `desc`) — ffx command.bin/
	// item.bin/a_ability.bin/monmagic1/2.bin.
	CommandLayout = LayoutSet{
		common.GameVersionFFX: {
			{"name", 0},
			{"name_simplified", 0},
			{"desc", 0},
			{"desc_simplified", 0},
		},
	}

	// KeyItemPairLayout: pares name + help — ffx important.bin/panel.bin
	// (o C# desses structs usa `help`, não `desc`).
	KeyItemPairLayout = LayoutSet{
		common.GameVersionFFX: {
			{"name", 0},
			{"name_simplified", 0},
			{"help", 0},
			{"help_simplified", 0},
		},
	}

	// CommandV2Layout: name + help (TextRef único) — ffx2/lastmiss
	// command/item/important/a_ability e lm_mes/lm_monster/lm_player/
	// lm_warehouse.
	CommandV2Layout = LayoutSet{
		common.GameVersionFFX2: {
			{"name", 0},
			{"help", 0},
		},
		common.GameVersionLastMiss: {
			{"name", 0},
			{"help", 0},
		},
	}

	// NameOnlyLayout: par name — ffx (arquivo sem segundo campo).
	NameOnlyLayout = LayoutSet{
		common.GameVersionFFX: {
			{"name", 0},
			{"name_simplified", 0},
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

	// MenuTextPairLayout: par command — ffx *_txt.bin de menu
	// (NameHelpText: command @0x00, command_simplified @0x04,
	// help @0x08, help_simplified @0x0C — confirmados no binário).
	MenuTextPairLayout = LayoutSet{
		common.GameVersionFFX: {
			{"command", 0},
			{"command_simplified", 0},
			{"help", 0},
			{"help_simplified", 0},
		},
	}

	// MenuTextV2Layout: command + help — ffx2 menu_txt.bin/btlend_txt.bin
	// (MenuTxt/BtlEndTxt: command@0x00, help@0x04).
	MenuTextV2Layout = LayoutSet{
		common.GameVersionFFX2: {
			{"command", 0},
			{"help", 0},
		},
	}

	// HelpTextV2Layout: help (TextRef) — ffx2 btl_txt.bin (BtlTxt).
	HelpTextV2Layout = LayoutSet{
		common.GameVersionFFX2: {
			{"help", 0},
		},
	}

	// HelpPairLayout: par help — ffx btl_txt.bin (HelpText) e sphere.bin
	// (Sphere: o único campo de texto é `help`).
	HelpPairLayout = LayoutSet{
		common.GameVersionFFX: {
			{"help", 0},
			{"help_simplified", 0},
		},
	}

	// PlyRomPairLayout: pares switch_text + scan_text — ffx ply_rom.bin
	// (PlyRom: SwitchText @0x00..0x08 e ScanText @0x08..0x10, confirmados
	// no binário).
	PlyRomPairLayout = LayoutSet{
		common.GameVersionFFX: {
			{"switch_text", 0},
			{"switch_text_simplified", 0},
			{"scan_text", 0},
			{"scan_text_simplified", 0},
		},
	}

	// AccessoryLayout: name + help + creature_data_help — ffx2 accessory.bin
	// (C#: name@0x00, help@0x04, creature_data.help@0x24).
	AccessoryLayout = LayoutSet{
		common.GameVersionFFX2: {
			{"name", 0},
			{"help", 0x1C},
			{"creature_data_help", 0},
		},
	}

	// JobLayout: name + help + creature_data_help — ffx2 job.bin
	// (C#: name@0x00, help@0x04, creature_data.help@0xAC).
	JobLayout = LayoutSet{
		common.GameVersionFFX2: {
			{"name", 0},
			{"help", 0xA4},
			{"creature_data_help", 0},
		},
	}

	// PlateLayout: name + help + messages_0..3 + creature_data_help —
	// ffx2 plate.bin (C#: name@0x00, help@0x04, messages@0x08..0x18,
	// creature_data.help@0x48). Os bytes 0x18..0x48 (bonus/icons/statchanges/
	// skill) são DADOS do struct — ficam no gap e voltam byte-a-byte no
	// ToBytes (clone do chunk original).
	PlateLayout = LayoutSet{
		common.GameVersionFFX2: {
			{"name", 0},
			{"help", 0},
			{"messages_0", 0},
			{"messages_1", 0},
			{"messages_2", 0},
			{"messages_3", 0x30},
			{"creature_data_help", 0},
		},
	}

	// MonsterLayout: name + pares sensor_text/scan_text — ffx monster1/2/3.bin
	// (MonStats: Name TextRef @0x00, SensorText @0x04..0x0C, ScanText @0x0C..0x14).
	MonsterLayout = LayoutSet{
		common.GameVersionFFX: {
			{"name", 0},
			{"sensor_text", 0},
			{"sensor_text_simplified", 0},
			{"scan_text", 0},
			{"scan_text_simplified", 0},
		},
	}

	// LastMissionLayout: name + help + effect + effect_description —
	// lm_accesary.bin (LmAccesary).
	LastMissionLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"help", 0},
			{"effect", 0},
			{"effect_description", 0},
		},
	}

	// LastMissionMesLayout: name + help contíguos, sem skip (LmMes).
	LastMissionMesLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"help", 0},
		},
	}

	// LastMissionDressLayout: name + help + effect (LmDress/LmTrap).
	LastMissionDressLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"help", 0},
			{"effect", 0},
		},
	}

	// LastMissionCommandLayout: name/help/information espaçados de 4
	// (LmCommand: name@0x00, help@0x08, information@0x10).
	LastMissionCommandLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 4},
			{"help", 4},
			{"information", 0},
		},
	}

	// LastMissionMonmagicLayout: name/help/information espaçados de 4
	// (LmMonMagic: name@0x00, help@0x08, information@0x10).
	LastMissionMonmagicLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 4},
			{"help", 4},
			{"information", 0},
		},
	}

	// LastMissionItemLayout: name/help/effect espaçados de 4 — lm_item.bin
	// (LmItem: name@0x00, help@0x08, effect@0x10).
	LastMissionItemLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 4},
			{"help", 4},
			{"effect", 0},
		},
	}

	LastMissionMonsterLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"help", 0},
		},
	}

	LastMissionPlayerLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"help", 0},
		},
	}

	LastMissionTrapLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"help", 0},
			{"effect", 0},
		},
	}

	LastMissionWarehouseLayout = LayoutSet{
		common.GameVersionLastMiss: {
			{"name", 0},
			{"help", 0},
		},
	}
)

// Formatters centralizados. São funções reutilizáveis de ToString,
// independentes do tipo concreto (posicionais sobre OrderedFieldKeys: as
// chaves variam por arquivo, a ordem binária não).
// O fallback do KeyedStringFile já é o join por " | " quando nenhum é definido.
var (
	// commandLegacyFmt une os 4 campos do par name + desc/para.
	commandLegacyFmt = func(f *KeyedStringFile, lang string) string {
		return fmt.Sprintf("%s %s - %s %s",
			FieldStringAt(f, 0, lang),
			FieldStringAt(f, 1, lang),
			FieldStringAt(f, 2, lang),
			FieldStringAt(f, 3, lang))
	}

	nameOnlyLegacyFmt = func(f *KeyedStringFile, lang string) string {
		return fmt.Sprintf("%s %s",
			FieldStringAt(f, 0, lang),
			FieldStringAt(f, 1, lang))
	}

	// threePartLegacyFmt une o primeiro, o segundo e o ÚLTIMO campo do
	// layout: os terceiros variam por arquivo (information/effect/
	// creature_data_help) mas o formato "nome - help - restante" é comum.
	threePartLegacyFmt = func(f *KeyedStringFile, lang string) string {
		keys := f.OrderedFieldKeys()
		third := keys[len(keys)-1]
		return fmt.Sprintf("%s - %s - %s",
			FieldString(f, keys[0], lang),
			FieldString(f, keys[1], lang),
			FieldString(f, third, lang))
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
