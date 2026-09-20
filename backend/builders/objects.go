package builders

import (
	"fmt"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/formatters/hash"
)

// objectsFilePatternPaths mapeia cada JSON de objectsfile exportado (nome base,
// sem sufixo de versão) para o pattern path do binário original, para que a
// metadata seja reconstruída. Movido de objectsfile (fonte única agora aqui).
var objectsFilePatternPaths = map[string]string{
	"key_items_all_localizations.json":       "battle/kernel/important.bin",
	"commands_all_localizations.json":        "battle/kernel/command.bin",
	"items_all_localizations.json":           "battle/kernel/item.bin",
	"a_ability_all_localizations.json":       "battle/kernel/a_ability.bin",
	"arms_text_all_localizations.json":       "battle/kernel/arms_txt.bin",
	"arms_all_localizations.json":            "battle/kernel/arms_txt.bin",
	"config_txt_all_localizations.json":      "battle/kernel/config_txt.bin",
	"config_text_all_localizations.json":     "battle/kernel/config_txt.bin",
	"item_txt_all_localizations.json":        "battle/kernel/item_txt.bin",
	"item_commands_all_localizations.json":   "battle/kernel/item_txt.bin",
	"mmain_txt_all_localizations.json":       "battle/kernel/mmain_txt.bin",
	"main_menu_all_localizations.json":       "battle/kernel/mmain_txt.bin",
	"player_rom_all_localizations.json":      "battle/kernel/ply_rom.bin",
	"player_room_all_localizations.json":     "battle/kernel/ply_rom.bin",
	"battle_text_all_localizations.json":     "battle/kernel/btl_txt.bin",
	"battle_end_text_all_localizations.json": "battle/kernel/btlend_txt.bin",
	"monster_magic1_all_localizations.json":  "battle/kernel/monmagic1.bin",
	"monster_magic2_all_localizations.json":  "battle/kernel/monmagic2.bin",
	"monster1_all_localizations.json":        "battle/kernel/monster1.bin",
	"monster2_all_localizations.json":        "battle/kernel/monster2.bin",
	"monster3_all_localizations.json":        "battle/kernel/monster3.bin",
	"build_txt_all_localizations.json":       "battle/kernel/build_txt.bin",
	"build_all_localizations.json":           "battle/kernel/build_txt.bin",
	"name_txt_all_localizations.json":        "battle/kernel/name_txt.bin",
	"names_all_localizations.json":           "battle/kernel/name_txt.bin",
	"panel_all_localizations.json":           "battle/kernel/panel.bin",
	"sphere_all_localizations.json":          "battle/kernel/sphere.bin",
	"save_txt_all_localizations.json":        "battle/kernel/save_txt.bin",
	"save_text_all_localizations.json":       "battle/kernel/save_txt.bin",
	"status_txt_all_localizations.json":      "battle/kernel/status_txt.bin",
	"summon_txt_all_localizations.json":      "battle/kernel/summon_txt.bin",
	"weapon_names_all_localizations.json":    "battle/kernel/w_name.bin",

	// FFX-2 (v2) kernel exclusivos.
	"accessory_all_localizations.json":     "battle/kernel/accessory.bin",
	"job_all_localizations.json":           "battle/kernel/job.bin",
	"menu_text_all_localizations.json":     "battle/kernel/menu_txt.bin",
	"monster_magic_all_localizations.json": "battle/kernel/monmagic.bin",
	"monster_all_localizations.json":       "battle/kernel/monster.bin",
	"oversoul_all_localizations.json":      "battle/kernel/oversoul.bin",
	"plate_all_localizations.json":         "battle/kernel/plate.bin",
	"player_save_all_localizations.json":   "battle/kernel/ply_save.bin",

	// LastMiss (lastmiss) kernel.
	"lm_accesary_text_all_localizations.json":  "lastmiss/kernel/lm_accesary.bin",
	"lm_command_text_all_localizations.json":   "lastmiss/kernel/lm_command.bin",
	"lm_dress_text_all_localizations.json":     "lastmiss/kernel/lm_dress.bin",
	"lm_floorname_text_all_localizations.json": "lastmiss/kernel/lm_floorname.bin",
	"lm_item_text_all_localizations.json":      "lastmiss/kernel/lm_item.bin",
	"lm_mes_text_all_localizations.json":       "lastmiss/kernel/lm_mes.bin",
	"lm_monmagic_text_all_localizations.json":  "lastmiss/kernel/lm_monmagic.bin",
	"lm_monster_text_all_localizations.json":   "lastmiss/kernel/lm_monster.bin",
	"lm_player_text_all_localizations.json":    "lastmiss/kernel/lm_player.bin",
	"lm_trap_text_all_localizations.json":      "lastmiss/kernel/lm_trap.bin",
	"lm_warehouse_text_all_localizations.json": "lastmiss/kernel/lm_warehouse.bin",
}

// PatternForJSONName resolve o pattern path do binário original a partir do
// nome do JSON (aceita com ou sem sufixo de versão).
func PatternForJSONName(fileName string) (string, bool) {
	if pattern, ok := objectsFilePatternPaths[fileName]; ok {
		return pattern, true
	}
	pattern, ok := objectsFilePatternPaths[common.StripVersionSuffix(fileName)]
	return pattern, ok
}

// ObjectsCollectionKey é a chave da Collection para um layout: o basename
// do binário sem extensão (nome do arquivo sem extensão, como nos demais
// formatos).
func ObjectsCollectionKey(layout objectsfile.FileLayout) string {
	return strings.TrimSuffix(layout.FileName, ".bin")
}

// BuildObjectsDTO monta a Collection pronta (texto + metadata + hash) a
// partir da lista de objetos em memória. Chave = basename sem extensão.
// Index das rows = posição do objeto na lista (reconstrução posicional).
func BuildObjectsDTO(objects components.IList[datastore.IGlobalLocalizedTextObject], layout objectsfile.FileLayout, key string) (dto.Collection, error) {
	if objects == nil || objects.IsEmpty() {
		return nil, fmt.Errorf("no objects loaded or empty")
	}
	entry := dto.FileEntry{
		Metadata: dto.NewObjectMetadata(layout.Version, layout.DirPattern, layout.FileName, key),
		Rows:     []dto.TextRow{},
	}
	objects.RangeIndex(func(i int, obj datastore.IGlobalLocalizedTextObject) {
		if obj == nil {
			common.LogVerbose("Object %d is nil, skipping", i)
			return
		}
		for _, f := range objectsfile.ExportFieldTexts(obj) {
			if len(f.Texts) == 0 {
				continue
			}
			entry.Rows = append(entry.Rows, dto.TextRow{
				Index: i,
				Name:  f.Key,
				Hash:  hash.Texts(f.Texts),
				Text:  f.Texts,
			})
		}
	})
	if len(entry.Rows) == 0 {
		return nil, fmt.Errorf("no objects with text data found")
	}
	dto.SortRows(entry.Rows)
	out := dto.Collection{ObjectsCollectionKey(layout): entry}
	// TODO: deletar quando colisão xxHash64 for considerada segura —
	// guarda temporária de desencargo: reprova DTO com mesmo hash para textos diferentes.
	if err := hash.ValidateNoCollision(out); err != nil {
		return nil, err
	}
	return out, nil
}

// ResolveObjectsLayout resolve layout e chave canônica a partir do nome
// legado do JSON e da versão (fluxos por nome de arquivo).
func ResolveObjectsLayout(version common.GameVersion, jsonFileName string) (objectsfile.FileLayout, string, error) {
	pattern, ok := PatternForJSONName(jsonFileName)
	if !ok {
		return objectsfile.FileLayout{}, "", fmt.Errorf("unknown objects JSON name: %s", jsonFileName)
	}
	layout, ok := objectsfile.FileLayoutFor(version, pattern)
	if !ok {
		return objectsfile.FileLayout{}, "", fmt.Errorf("no layout registered for %s version %s", pattern, version)
	}
	return layout, objectsfile.FileLayoutKey(version, pattern), nil
}

// BuildObjectsDTOByJSONName é como BuildObjectsDTO, mas resolvendo
// layout/chave a partir do nome legado do JSON (fluxos por nome de arquivo).
func BuildObjectsDTOByJSONName(objects components.IList[datastore.IGlobalLocalizedTextObject], version common.GameVersion, jsonFileName string) (dto.Collection, error) {
	layout, key, err := ResolveObjectsLayout(version, jsonFileName)
	if err != nil {
		return nil, err
	}
	return BuildObjectsDTO(objects, layout, key)
}

// ApplyObjectsEntry aplica uma entrada do DTO de volta na lista em memória
// (parse DTO → objetos). Não persiste em disco: quem salva é SaveToBinary.
// A metadata Key, quando presente, precisa bater com a chave esperada.
func ApplyObjectsEntry(objects components.IList[datastore.IGlobalLocalizedTextObject], version common.GameVersion, key string, entry dto.FileEntry) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}
	if entry.Metadata.Key != "" && entry.Metadata.Key != key {
		return fmt.Errorf("metadata key mismatch: DTO %q vs expected %q", entry.Metadata.Key, key)
	}
	items := objects.Items()
	dto.SortRows(entry.Rows)
	for _, row := range entry.Rows {
		if row.Index < 0 || row.Index >= len(items) {
			common.LogError("Object ID without range: %d", row.Index)
			continue
		}
		obj := items[row.Index]
		if obj == nil {
			common.LogError("Localized object not found: %d", row.Index)
			continue
		}
		for lang, newText := range row.Text {
			if want, ok := row.Hash[lang]; ok && want != "" && newText != "" {
				if got := hash.Sum64Hex(newText); got != want {
					common.LogVerbose("hash mismatch for object %d field %q lang %s: file %s vs text %s",
						row.Index, row.Name, lang, want, got)
				}
			}
		}
		common.LogVerbose("Processing localized object %d field %q", row.Index, row.Name)
		objectsfile.ApplyFieldTexts(obj, []objectsfile.FieldText{{Key: row.Name, Texts: row.Text}}, version)
	}
	common.LogVerbose("Localized objects updated successfully (%d rows)", len(entry.Rows))
	return nil
}
