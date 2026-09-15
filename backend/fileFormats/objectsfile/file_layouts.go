package objectsfile

import "ffxresources/backend/common"

// FileLayout declarativo especifica como carregar/decodificar um arquivo binário de objetos.
// Cada layout é único por (Version, DirPattern, FileName) — arquivos como "command.bin"
// existem em FFX e FFX2 com campos diferentes.
//
// Fields/Formatter is ignored when Creator is set (e.g. weapons); otherwise a generic
// KeyedStringFile creator is built from LayoutSet{Version: Fields}.
// Start is the byte offset into the object chunk where string segments begin (0 for most).
type FileLayout struct {
	Version    common.GameVersion
	DirPattern string
	FileName   string
	IndexCount int
	Fields     []SegmentField
	Start      int
	Formatter  StringFormatter
	Creator    CreatorFunc
}

func (l FileLayout) PatternPath() string { return l.DirPattern + "/" + l.FileName }

func FileLayoutKey(version common.GameVersion, patternPath string) string {
	return version.String() + "/" + patternPath
}

func FileLayoutFor(version common.GameVersion, patternPath string) (FileLayout, bool) {
	key := FileLayoutKey(version, patternPath)
	layout, ok := FileLayouts[key]
	return layout, ok
}

// FileLayouts mapeia "version/patternPath" -> FileLayout.
var FileLayouts = map[string]FileLayout{
	// ===== FFX (v1) - battle/kernel =====
	// Command layout (name, simplifiedName, description, simplifiedDescription)
	"ffx/battle/kernel/command.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "command.bin",
		IndexCount: 318,
		Fields:     CommandLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  commandLegacyFmt,
	},
	"ffx/battle/kernel/important.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "important.bin",
		IndexCount: 51,
		Fields:     CommandLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  commandLegacyFmt,
	},
	"ffx/battle/kernel/item.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "item.bin",
		IndexCount: 155,
		Fields:     CommandLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  commandLegacyFmt,
	},
	"ffx/battle/kernel/a_ability.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "a_ability.bin",
		IndexCount: 32,
		Fields:     CommandLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  commandLegacyFmt,
	},
	"ffx/battle/kernel/panel.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "panel.bin",
		IndexCount: 79,
		Fields:     CommandLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  commandLegacyFmt,
	},
	// NameOnly layout (name, simplifiedName)
	"ffx/battle/kernel/arms_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "arms_txt.bin",
		IndexCount: 9,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/config_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "config_txt.bin",
		IndexCount: 13,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/item_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "item_txt.bin",
		IndexCount: 7,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/mmain_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "mmain_txt.bin",
		IndexCount: 19,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/menu_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "menu_txt.bin",
		IndexCount: 6,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/summon_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "summon_txt.bin",
		IndexCount: 5,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/status_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "status_txt.bin",
		IndexCount: 9,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/btl_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "btl_txt.bin",
		IndexCount: 12,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/btlend_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "btlend_txt.bin",
		IndexCount: 2,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/monmagic1.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monmagic1.bin",
		IndexCount: 0,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/monmagic2.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monmagic2.bin",
		IndexCount: 0,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/build_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "build_txt.bin",
		IndexCount: 0,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/name_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "name_txt.bin",
		IndexCount: 109,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/monster1.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monster1.bin",
		IndexCount: 69,
		Fields:     MonsterLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nil,
	},
	"ffx/battle/kernel/monster2.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monster2.bin",
		IndexCount: 23,
		Fields:     MonsterLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nil,
	},
	"ffx/battle/kernel/monster3.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monster3.bin",
		IndexCount: 83,
		Fields:     MonsterLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nil,
	},
	"ffx/battle/kernel/ply_rom.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "ply_rom.bin",
		IndexCount: 0,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/ply_save.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "ply_save.bin",
		IndexCount: 1,
		Fields:     NameOnlyV2Layout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nil,
	},
	"ffx/battle/kernel/sphere.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "sphere.bin",
		IndexCount: 6,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	"ffx/battle/kernel/save_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "save_txt.bin",
		IndexCount: 6,
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Start:      0,
		Formatter:  nameOnlyLegacyFmt,
	},
	// Weapon names - special creator (excluded from generic layouts)
	// TODO: w_name.bin (WeaponsNameTextObject) still needs adaptation to FileLayout format
	// due to its unique struct layout (array of Tidus/Yuna/Auron/etc. names)
	// "ffx/battle/kernel/w_name.bin": {
	// 	Version:    common.GameVersionFFX,
	// 	DirPattern: "battle/kernel",
	// 	FileName:   "w_name.bin",
	// 	IndexCount: 7,
	// 	Creator:    NewWeaponsNameTextObject,
	// 	Start:      0,
	// },

	// ===== FFX-2 (v2) - battle/kernel =====
	// CommandV2 layout (name, description)
	"ffx2/battle/kernel/command.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "command.bin",
		IndexCount: 0,
		Fields:     CommandV2Layout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  nil,
	},
	"ffx2/battle/kernel/item.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "item.bin",
		IndexCount: 0,
		Fields:     CommandV2Layout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  nil,
	},
	"ffx2/battle/kernel/important.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "important.bin",
		IndexCount: 0,
		Fields:     CommandV2Layout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  nil,
	},
	// Accessory/Job/Plate layouts (three-part with effect)
	"ffx2/battle/kernel/accessory.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "accessory.bin",
		IndexCount: 0,
		Fields:     AccessoryLayout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  threePartLegacyFmt,
	},
	"ffx2/battle/kernel/job.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "job.bin",
		IndexCount: 0,
		Fields:     JobLayout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  threePartLegacyFmt,
	},
	"ffx2/battle/kernel/plate.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "plate.bin",
		IndexCount: 0,
		Fields:     PlateLayout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  threePartLegacyFmt,
	},
	// NameOnlyV2 layout
	"ffx2/battle/kernel/a_ability.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "a_ability.bin",
		IndexCount: 0,
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  nil,
	},
	"ffx2/battle/kernel/menu_txt.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "menu_txt.bin",
		IndexCount: 0,
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  nil,
	},
	"ffx2/battle/kernel/monmagic.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "monmagic.bin",
		IndexCount: 0,
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  nil,
	},
	"ffx2/battle/kernel/monster.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "monster.bin",
		IndexCount: 0,
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  nil,
	},
	"ffx2/battle/kernel/monster2.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "monster2.bin",
		IndexCount: 0,
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  nil,
	},
	"ffx2/battle/kernel/oversoul.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "oversoul.bin",
		IndexCount: 0,
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  nil,
	},
	"ffx2/battle/kernel/btl_txt.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "btl_txt.bin",
		IndexCount: 0,
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  nil,
	},
	"ffx2/battle/kernel/btlend_txt.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "btlend_txt.bin",
		IndexCount: 0,
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Start:      0,
		Formatter:  nil,
	},

	// ===== LastMission (lastmiss) =====
	// LastMission layout (name, description, effect, effectDescription)
	"lastmiss/kernel/lm_accesary.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_accesary.bin",
		IndexCount: 0,
		Fields:     LastMissionLayout[common.GameVersionLastMiss],
		Start:      0,
		Formatter:  nil,
	},
	// LastMission command layout (name, description, effect - with skip positional)
	"lastmiss/kernel/lm_command.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_command.bin",
		IndexCount: 0,
		Fields:     LastMissionCommandLayout[common.GameVersionLastMiss],
		Start:      4,
		Formatter:  threePartLegacyFmt,
	},
	// lm_item uses the same skip-layout as lm_command
	"lastmiss/kernel/lm_item.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_item.bin",
		IndexCount: 0,
		Fields:     LastMissionCommandLayout[common.GameVersionLastMiss],
		Start:      4,
		Formatter:  threePartLegacyFmt,
	},
	// LastMission dress layout (name, description, effect)
	"lastmiss/kernel/lm_dress.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_dress.bin",
		IndexCount: 0,
		Fields:     LastMissionDressLayout[common.GameVersionLastMiss],
		Start:      0,
		Formatter:  threePartLegacyFmt,
	},
	// LastMission Mes layout (name + description contiguous, no skip)
	"lastmiss/kernel/lm_mes.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_mes.bin",
		IndexCount: 0,
		Fields:     LastMissionMesLayout[common.GameVersionLastMiss],
		Start:      0,
		Formatter:  nil,
	},
	// LastMission Monmagic layout (name/description skip 4, effect no skip)
	"lastmiss/kernel/lm_monmagic.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_monmagic.bin",
		IndexCount: 0,
		Fields:     LastMissionMonmagicLayout[common.GameVersionLastMiss],
		Start:      4,
		Formatter:  threePartLegacyFmt,
	},
	// LastMission Monster layout (name + description, no skip)
	"lastmiss/kernel/lm_monster.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_monster.bin",
		IndexCount: 0,
		Fields:     LastMissionMonsterLayout[common.GameVersionLastMiss],
		Start:      0,
		Formatter:  nil,
	},
	// LastMission Player layout (name + description, no skip)
	"lastmiss/kernel/lm_player.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_player.bin",
		IndexCount: 0,
		Fields:     LastMissionPlayerLayout[common.GameVersionLastMiss],
		Start:      0,
		Formatter:  nil,
	},
	// LastMission Trap layout (name, description, effect, no skip)
	"lastmiss/kernel/lm_trap.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_trap.bin",
		IndexCount: 0,
		Fields:     LastMissionTrapLayout[common.GameVersionLastMiss],
		Start:      0,
		Formatter:  threePartLegacyFmt,
	},
	// LastMission Warehouse layout (name + description, no skip)
	"lastmiss/kernel/lm_warehouse.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_warehouse.bin",
		IndexCount: 0,
		Fields:     LastMissionWarehouseLayout[common.GameVersionLastMiss],
		Start:      0,
		Formatter:  nil,
	},
	// LastMission Floorname layout (NameOnlyV2)
	"lastmiss/kernel/lm_floorname.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_floorname.bin",
		IndexCount: 0,
		Fields:     NameOnlyV2Layout[common.GameVersionLastMiss],
		Start:      0,
		Formatter:  nil,
	},
}
