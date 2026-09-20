package objectsfile

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

func fillLocalized(m map[string]string, seg datastore.IGlobalLocalizedKeyedStringObject, locKey string) map[string]string {
	if seg == nil {
		return m
	}
	if text := seg.GetLocalizedString(locKey); text != "" {
		if m == nil {
			m = make(map[string]string)
		}
		m[locKey] = text
	}
	return m
}

type binding struct {
	seg    datastore.IGlobalLocalizedKeyedStringObject
	target *map[string]string
}

func bind(obj datastore.IGlobalLocalizedTextObject, key string, target *map[string]string) binding {
	return binding{seg: obj.GetKeyedString(key), target: target}
}

func (b binding) export(locKey string) {
	if b.seg == nil {
		return
	}
	if text := b.seg.GetLocalizedString(locKey); text != "" {
		if *b.target == nil {
			*b.target = make(map[string]string)
		}
		(*b.target)[locKey] = text
	}
}

func abilityBindings(obj datastore.IGlobalLocalizedTextObject, data *datastore.ObjectTextEntry) []binding {
	var segs []datastore.IGlobalLocalizedKeyedStringObject
	for i := 1; ; i++ {
		seg := obj.GetKeyedString("ability" + strconv.Itoa(i))
		if seg == nil {
			break
		}
		segs = append(segs, seg)
	}
	if len(segs) == 0 {
		return nil
	}
	data.Abilities = make([]map[string]string, len(segs))
	bs := make([]binding, len(segs))
	for i, seg := range segs {
		data.Abilities[i] = make(map[string]string)
		bs[i] = binding{seg: seg, target: &data.Abilities[i]}
	}
	return bs
}

type weaponsTexts struct {
	weapon *WeaponsNameTextObject
}

func resolveWeaponsTexts(obj datastore.IGlobalLocalizedTextObject) (weaponsTexts, bool) {
	w, ok := obj.(*WeaponsNameTextObject)
	return weaponsTexts{weapon: w}, ok
}

func (w weaponsTexts) exportTo(data *datastore.ObjectTextEntry, locKey string) {
	if data.Weapons == nil {
		data.Weapons = make(map[string]datastore.WeaponTexts, len(weaponRefs))
	}
	for i, ref := range weaponRefs {
		entry := data.Weapons[ref.name]
		if entry.Name == nil {
			entry.Name = make(map[string]string)
			entry.SimplifiedName = make(map[string]string)
		}
		entry.Name = fillLocalized(entry.Name, w.weapon.Names[i], locKey)
		entry.SimplifiedName = fillLocalized(entry.SimplifiedName, w.weapon.SimplifiedNames[i], locKey)
		data.Weapons[ref.name] = entry
	}
}

func buildJSONEntries(objects components.IList[datastore.IGlobalLocalizedTextObject]) []*datastore.ObjectTextEntry {
	var jsonEntries []*datastore.ObjectTextEntry
	objects.RangeIndex(func(i int, obj datastore.IGlobalLocalizedTextObject) {
		if obj == nil {
			common.LogVerbose("Object %d is nil, skipping", i)
			return
		}

		data := &datastore.ObjectTextEntry{ID: i}

		var bs []binding
		bs = append(bs, staticBindings(obj, data)...)
		if ab := abilityBindings(obj, data); ab != nil {
			bs = append(bs, ab...)
		}

		if w, ok := resolveWeaponsTexts(obj); ok {
			for locKey := range common.SupportedLanguages {
				w.exportTo(data, locKey)
			}
		}

		for locKey := range common.SupportedLanguages {
			for _, b := range bs {
				b.export(locKey)
			}
		}

		if data.HasContent() {
			jsonEntries = append(jsonEntries, data)
		}
	})
	return jsonEntries
}

func ExportToJSON(objects components.IList[datastore.IGlobalLocalizedTextObject], jsonFileName string, formatter datastore.IObjectsFormatter) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}
	return createJSON(buildJSONEntries(objects), jsonFileName, formatter)
}

// objectsFilePatternPaths maps each exported objectsfile JSON (nome base,
// sem sufixo de versão) to the pattern path of the original game binary
// it was extracted from, so its metadata can be reconstructed.
// Mantém aliases legados + todos os nomes usados em
// examples/runFFXExamples, runFFX2Examples e runLastMissExamples.
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

// binaryMetadataForFile returns metadata describing the original binary for a given
// objectsfile export filename (aceita nome com ou sem sufixo de versão),
// ou nil quando o mapeamento é desconhecido.
func binaryMetadataForFile(fileName string) *models.FileMetadata {
	pattern, ok := objectsFilePatternPaths[fileName]
	if !ok {
		pattern, ok = objectsFilePatternPaths[common.StripVersionSuffix(fileName)]
	}
	if !ok {
		return nil
	}
	return models.NewFileMetadata(models.NewFileInfoFromPath(models.ObjectFileBinaryPath(pattern)))
}

func createJSON(dataObjectEntries []*datastore.ObjectTextEntry, fileName string, formatter datastore.IObjectsFormatter) error {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	jsonPath := filepath.Join(editsPath, common.WithVersionSuffix(fileName))

	raw, err := formatter.Marshal(datastore.ObjectTextData{
		Entries:  dataObjectEntries,
		Metadata: binaryMetadataForFile(fileName),
	})
	if err != nil {
		return fmt.Errorf("error marshaling data for %s: %w", jsonPath, err)
	}
	if err := common.WriteBytesToFile(jsonPath, raw); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", jsonPath, err)
	}

	common.LogVerbose("Exported name-description data to JSON: %s", jsonPath)
	return nil
}

// ExportToJSONForStore exports objects carrying the keyed layout metadata,
// without depending on objectsFilePatternPaths.
func ExportToJSONForStore(objects components.IList[datastore.IGlobalLocalizedTextObject], layout FileLayout, key string, formatter datastore.IObjectsFormatter) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}
	return createJSONForStore(buildJSONEntries(objects), layout, key, formatter)
}

func createJSONForStore(dataObjectEntries []*datastore.ObjectTextEntry, layout FileLayout, key string, formatter datastore.IObjectsFormatter) error {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}
	fileName := strings.ReplaceAll(strings.ReplaceAll(key, "/", "_"), ".bin", ".json")
	jsonPath := filepath.Join(editsPath, fileName)

	raw, err := formatter.Marshal(datastore.ObjectTextData{
		Entries:  dataObjectEntries,
		Metadata: models.NewObjectFileMetadataKeyed(layout.Version, layout.DirPattern, layout.FileName, key),
	})
	if err != nil {
		return fmt.Errorf("error marshaling data for %s: %w", jsonPath, err)
	}
	if err := common.WriteBytesToFile(jsonPath, raw); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", jsonPath, err)
	}

	common.LogVerbose("Exported name-description data to JSON: %s", jsonPath)
	return nil
}
