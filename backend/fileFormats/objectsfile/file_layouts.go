package objectsfile

import (
	"strings"

	"ffxresources/backend/common"
)

// FileLayout declara a identidade de um arquivo de objetos: versão, diretório
// e nome. A leitura/gravação pertence ao fluxo chunkmap
// (backend/fileFormats/objectsfile/chunkmap); este mapa serve ao
// resolveObjectLayouts/objectKeyForID (iteração e resolução de layouts por id).
type FileLayout struct {
	Version    common.GameVersion
	DirPattern string
	FileName   string
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
	},
	"ffx/battle/kernel/important.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "important.bin",
	},
	"ffx/battle/kernel/item.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "item.bin",
	},
	"ffx/battle/kernel/a_ability.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "a_ability.bin",
	},
	"ffx/battle/kernel/panel.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "panel.bin",
	},
	"ffx/battle/kernel/arms_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "arms_txt.bin",
	},
	"ffx/battle/kernel/config_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "config_txt.bin",
	},
	"ffx/battle/kernel/item_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "item_txt.bin",
	},
	"ffx/battle/kernel/mmain_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "mmain_txt.bin",
	},
	"ffx/battle/kernel/menu_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "menu_txt.bin",
	},
	"ffx/battle/kernel/summon_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "summon_txt.bin",
	},
	"ffx/battle/kernel/status_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "status_txt.bin",
	},
	"ffx/battle/kernel/btl_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "btl_txt.bin",
	},
	"ffx/battle/kernel/btlend_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "btlend_txt.bin",
	},
	// monmagic1/2.bin: o C# (FFX.Command) tem os 4 campos de texto
	// (name + desc) — confirmados no binário; antes lia só o par name.
	"ffx/battle/kernel/monmagic1.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monmagic1.bin",
	},
	"ffx/battle/kernel/monmagic2.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monmagic2.bin",
	},
	"ffx/battle/kernel/build_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "build_txt.bin",
	},
	"ffx/battle/kernel/name_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "name_txt.bin",
	},
	"ffx/battle/kernel/monster1.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monster1.bin",
	},
	"ffx/battle/kernel/monster2.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monster2.bin",
	},
	"ffx/battle/kernel/monster3.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "monster3.bin",
	},
	"ffx/battle/kernel/ply_rom.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "ply_rom.bin",
	},

	"ffx/battle/kernel/ply_save.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "ply_save.bin",
	},
	"ffx/battle/kernel/sphere.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "sphere.bin",
	},
	"ffx/battle/kernel/save_txt.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "save_txt.bin",
	},
	// w_name.bin — weapon names (FFX). DEPRECATED: o fluxo chunkmap
	// (SchemaRegistry["ffx/battle/kernel/w_name.bin"]) substitui este Creator.
	// Mantido apenas para resolveObjectLayouts e objectKeyForID funcionarem,
	// e como fallback para pipeline de import de JSON antigo.
	"ffx/battle/kernel/w_name.bin": {
		Version:    common.GameVersionFFX,
		DirPattern: "battle/kernel",
		FileName:   "w_name.bin",
	},

	// ===== FFX-2 (v2) - battle/kernel =====
	"ffx2/battle/kernel/command.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "command.bin",
	},
	"ffx2/battle/kernel/item.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "item.bin",
	},
	"ffx2/battle/kernel/important.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "important.bin",
	},
	"ffx2/battle/kernel/accessory.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "accessory.bin",
	},
	"ffx2/battle/kernel/job.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "job.bin",
	},
	"ffx2/battle/kernel/plate.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "plate.bin",
	},
	"ffx2/battle/kernel/a_ability.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "a_ability.bin",
	},
	"ffx2/battle/kernel/menu_txt.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "menu_txt.bin",
	},
	// monmagic/monster/monster2: name + help (MCommand/Monster/Monster2 —
	// confirmados no binário; antes lia só o name).
	"ffx2/battle/kernel/monmagic.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "monmagic.bin",
	},
	"ffx2/battle/kernel/monster.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "monster.bin",
	},
	"ffx2/battle/kernel/monster2.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "monster2.bin",
	},
	"ffx2/battle/kernel/oversoul.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "oversoul.bin",
	},
	"ffx2/battle/kernel/btl_txt.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "btl_txt.bin",
	},
	"ffx2/battle/kernel/btlend_txt.bin": {
		Version:    common.GameVersionFFX2,
		DirPattern: "battle/kernel",
		FileName:   "btlend_txt.bin",
	},

	// ===== LastMission (lastmiss) =====
	"lastmiss/kernel/lm_accesary.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_accesary.bin",
	},
	"lastmiss/kernel/lm_command.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_command.bin",
	},
	"lastmiss/kernel/lm_item.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_item.bin",
	},
	"lastmiss/kernel/lm_dress.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_dress.bin",
	},
	"lastmiss/kernel/lm_mes.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_mes.bin",
	},
	"lastmiss/kernel/lm_monmagic.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_monmagic.bin",
	},
	"lastmiss/kernel/lm_monster.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_monster.bin",
	},
	"lastmiss/kernel/lm_player.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_player.bin",
	},
	"lastmiss/kernel/lm_trap.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_trap.bin",
	},
	"lastmiss/kernel/lm_warehouse.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_warehouse.bin",
	},
	"lastmiss/kernel/lm_floorname.bin": {
		Version:    common.GameVersionLastMiss,
		DirPattern: "lastmiss/kernel",
		FileName:   "lm_floorname.bin",
	},
}
