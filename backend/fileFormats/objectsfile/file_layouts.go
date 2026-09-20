package objectsfile

import (
	"strings"

	"ffxresources/backend/common"
)

// FileLayout declarativo especifica como carregar/decodificar um arquivo binário de objetos.
// Cada layout é único por (Version, DirPattern, FileName) — arquivos como "command.bin"
// existem em FFX e FFX2 com campos diferentes. Cada entrada carrega seus Fields
// inline explícitos (repetição intencional para ajuste posterior por arquivo).
//
// Fields/Formatter é ignorado quando Creator está definido (ex. weapons); caso
// contrário um creator genérico KeyedStringFileStore é construído a partir de
// LayoutSet{Version: Fields}. O primeiro segmento sempre começa em 0; saltos
// são expressos via Gap (0 = contíguo, >0 = bytes pulados após o segmento).
type FileLayout struct {
	Version    common.GameVersion
	DirPattern string
	FileName   string
	Fields     []SegmentField
	Formatter  StringFormatterStore
	Creator    CreatorFunc
}

func (l FileLayout) PatternPath() string { return l.DirPattern + "/" + l.FileName }

// FilePath retorna o caminho do arquivo com a raiz de versão do layout:
// "ffx/..." para FFX, "ffx2/..." para FFX2 e LastMiss (expansão sob ffx2).
func (l FileLayout) FilePath() string {
	return common.VersionPathName(l.Version) + "/" + l.PatternPath()
}

func FileLayoutKey(version common.GameVersion, patternPath string) string {
	prefix := version.String() + "/"
	if strings.HasPrefix(patternPath, prefix) {
		return patternPath
	}
	return prefix + patternPath
}

func FileLayoutFor(version common.GameVersion, patternPath string) (FileLayout, bool) {
	key := FileLayoutKey(version, patternPath)
	layout, ok := FileLayouts[key]
	return layout, ok
}

// FileLayouts mapeia "version/patternPath" -> FileLayout.
var FileLayouts = map[string]FileLayout{
	// ===== FFX (v1) - battle/kernel =====
	"ffx/battle/kernel/command.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "command.bin",
		Fields:     CommandLayout[common.GameVersionFFX],
		Formatter:  commandLegacyFmtStore,
	},
	"ffx/battle/kernel/important.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "important.bin",
		Fields:     CommandLayout[common.GameVersionFFX],
		Formatter:  commandLegacyFmtStore,
	},
	"ffx/battle/kernel/item.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "item.bin",
		Fields:     CommandLayout[common.GameVersionFFX],
		Formatter:  commandLegacyFmtStore,
	},
	"ffx/battle/kernel/a_ability.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "a_ability.bin",
		Fields:     CommandLayout[common.GameVersionFFX],
		Formatter:  commandLegacyFmtStore,
	},
	"ffx/battle/kernel/panel.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "panel.bin",
		Fields:     CommandLayout[common.GameVersionFFX],
		Formatter:  commandLegacyFmtStore,
	},
	"ffx/battle/kernel/arms_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "arms_txt.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/config_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "config_txt.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/item_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "item_txt.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/mmain_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "mmain_txt.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/menu_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "menu_txt.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/summon_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "summon_txt.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/status_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "status_txt.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/btl_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "btl_txt.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/btlend_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "btlend_txt.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/monmagic1.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monmagic1.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/monmagic2.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monmagic2.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/build_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "build_txt.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/name_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "name_txt.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/monster1.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monster1.bin",
		Fields:     MonsterLayout[common.GameVersionFFX],
		Formatter:  nil,
	},
	"ffx/battle/kernel/monster2.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monster2.bin",
		Fields:     MonsterLayout[common.GameVersionFFX],
		Formatter:  nil,
	},
	"ffx/battle/kernel/monster3.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monster3.bin",
		Fields:     MonsterLayout[common.GameVersionFFX],
		Formatter:  nil,
	},
	"ffx/battle/kernel/ply_rom.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "ply_rom.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/ply_save.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "ply_save.bin",
		Fields:     NameOnlyV2Layout[common.GameVersionFFX],
		Formatter:  nil,
	},
	"ffx/battle/kernel/sphere.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "sphere.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	"ffx/battle/kernel/save_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "save_txt.bin",
		Fields:     NameOnlyLayout[common.GameVersionFFX],
		Formatter:  nameOnlyLegacyFmtStore,
	},
	// Weapon names - special creator (excluded from generic layouts)
	// TODO: w_name.bin (WeaponsNameTextObject) still needs adaptation to FileLayout format
	// due to its unique struct layout (array of Tidus/Yuna/Auron/etc. names)

	// ===== FFX-2 (v2) - battle/kernel =====
	"ffx2/battle/kernel/command.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "command.bin",
		Fields:     CommandV2Layout[common.GameVersionFFX2],
		Formatter:  nil,
	},
	"ffx2/battle/kernel/item.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "item.bin",
		Fields:     CommandV2Layout[common.GameVersionFFX2],
		Formatter:  nil,
	},
	"ffx2/battle/kernel/important.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "important.bin",
		Fields:     CommandV2Layout[common.GameVersionFFX2],
		Formatter:  nil,
	},
	"ffx2/battle/kernel/accessory.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "accessory.bin",
		Fields: []SegmentField{
			{"name", 0},
			{"description", 0x1C},
			{"effect", 0},
		},
		Formatter: threePartLegacyFmtStore,
	},
	"ffx2/battle/kernel/job.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "job.bin",
		Fields: []SegmentField{
			{"name", 0},
			{"description", 0xA4},
			{"effect", 0},
		},
		Formatter: threePartLegacyFmtStore,
	},
	"ffx2/battle/kernel/plate.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "plate.bin",
		Fields:     PlateLayout[common.GameVersionFFX2],
		Formatter:  threePartLegacyFmtStore,
	},
	"ffx2/battle/kernel/a_ability.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "a_ability.bin",
		Fields:     CommandV2Layout[common.GameVersionFFX2],
		Formatter:  nil,
	},
	"ffx2/battle/kernel/menu_txt.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "menu_txt.bin",
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Formatter:  nil,
	},
	"ffx2/battle/kernel/monmagic.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "monmagic.bin",
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Formatter:  nil,
	},
	"ffx2/battle/kernel/monster.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "monster.bin",
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Formatter:  nil,
	},
	"ffx2/battle/kernel/monster2.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "monster2.bin",
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Formatter:  nil,
	},
	"ffx2/battle/kernel/oversoul.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "oversoul.bin",
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Formatter:  nil,
	},
	"ffx2/battle/kernel/btl_txt.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "btl_txt.bin",
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Formatter:  nil,
	},
	"ffx2/battle/kernel/btlend_txt.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "btlend_txt.bin",
		Fields:     NameOnlyV2Layout[common.GameVersionFFX2],
		Formatter:  nil,
	},

	// ===== LastMission (lastmiss) =====
	"lastmiss/kernel/lm_accesary.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_accesary.bin",
		Fields: []SegmentField{
			{"name", 0},
			{"description", 0},
			{"effect", 0},
			{"effectDescription", 0},
		},
		Formatter: nil,
	},
	"lastmiss/kernel/lm_command.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_command.bin",
		Fields: []SegmentField{
			{"name", 4},
			{"description", 4},
			{"effect", 0},
		},
		Formatter: threePartLegacyFmtStore,
	},
	"lastmiss/kernel/lm_item.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_item.bin",
		Fields:     LastMissionMonmagicLayout[common.GameVersionLastMiss],
		Formatter:  threePartLegacyFmtStore,
	},
	"lastmiss/kernel/lm_dress.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_dress.bin",
		Fields:     LastMissionDressLayout[common.GameVersionLastMiss],
		Formatter:  threePartLegacyFmtStore,
	},
	"lastmiss/kernel/lm_mes.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_mes.bin",
		Fields:     CommandV2Layout[common.GameVersionLastMiss],
		Formatter:  nil,
	},
	"lastmiss/kernel/lm_monmagic.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_monmagic.bin",
		Fields:     LastMissionMonmagicLayout[common.GameVersionLastMiss],
		Formatter:  threePartLegacyFmtStore,
	},
	"lastmiss/kernel/lm_monster.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_monster.bin",
		Fields:     CommandV2Layout[common.GameVersionLastMiss],
		Formatter:  nil,
	},
	"lastmiss/kernel/lm_player.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_player.bin",
		Fields:     CommandV2Layout[common.GameVersionLastMiss],
		Formatter:  nil,
	},
	"lastmiss/kernel/lm_trap.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_trap.bin",
		Fields:     LastMissionDressLayout[common.GameVersionLastMiss],
		Formatter:  threePartLegacyFmtStore,
	},
	"lastmiss/kernel/lm_warehouse.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_warehouse.bin",
		Fields:     CommandV2Layout[common.GameVersionLastMiss],
		Formatter:  nil,
	},
	"lastmiss/kernel/lm_floorname.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_floorname.bin",
		Fields:     NameOnlyV2Layout[common.GameVersionLastMiss],
		Formatter:  nil,
	},
}
