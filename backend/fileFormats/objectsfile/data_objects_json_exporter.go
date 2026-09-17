package objectsfile

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/interactions"
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

func abilityBindings(obj datastore.IGlobalLocalizedTextObject, data *JSONEntry) []binding {
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

func (w weaponsTexts) exportTo(data *JSONEntry, locKey string) {
	if data.Weapons == nil {
		data.Weapons = make(map[string]WeaponTexts, len(weaponRefs))
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

func (e *JSONEntry) hasContent() bool {
	if len(e.Name) > 0 || len(e.SimplifiedName) > 0 ||
		len(e.Description) > 0 || len(e.SimplifiedDescription) > 0 ||
		len(e.Effect) > 0 || len(e.EffectDescription) > 0 ||
		len(e.Bonus) > 0 || len(e.BonusIconA) > 0 ||
		len(e.BonusIconB) > 0 || len(e.BonusReserve) > 0 ||
		len(e.SensorText) > 0 || len(e.SimplifiedSensorText) > 0 ||
		len(e.ScanText) > 0 || len(e.SimplifiedScanText) > 0 {
		return true
	}
	for _, ab := range e.Abilities {
		if len(ab) > 0 {
			return true
		}
	}
	for _, wt := range e.Weapons {
		if len(wt.Name) > 0 || len(wt.SimplifiedName) > 0 {
			return true
		}
	}
	return false
}

func buildJSONEntries(objects components.IList[datastore.IGlobalLocalizedTextObject]) []*JSONEntry {
	var jsonEntries []*JSONEntry
	objects.RangeIndex(func(i int, obj datastore.IGlobalLocalizedTextObject) {
		if obj == nil {
			common.LogVerbose("Object %d is nil, skipping", i)
			return
		}

		data := &JSONEntry{ID: i}

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

		if data.hasContent() {
			jsonEntries = append(jsonEntries, data)
		}
	})
	return jsonEntries
}

func ExportToJSON(objects components.IList[datastore.IGlobalLocalizedTextObject], jsonFileName string) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}
	return createJSON(buildJSONEntries(objects), jsonFileName)
}

// EventFileDataJSON represents an event file with all its strings for JSON export
type EventFileDataJSON struct {
	ID      string                `json:"id"`
	Strings []EventStringDataJSON `json:"strings"`
}

// EventStringDataJSON represents a single event string with its localizations
type EventStringDataJSON struct {
	Index int               `json:"index"`
	Text  map[string]string `json:"text"`
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

func createJSON(dataObjectEntries []*JSONEntry, fileName string) error {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	jsonPath := filepath.Join(editsPath, common.WithVersionSuffix(fileName))

	stringsBytes, err := json.Marshal(dataObjectEntries)
	if err != nil {
		return fmt.Errorf("error marshaling data for %s: %w", jsonPath, err)
	}

	export := models.ObjectsFileExport{
		Metadata: binaryMetadataForFile(fileName),
		Strings:  stringsBytes,
	}

	if err := models.SaveDataFile(export, jsonPath); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", jsonPath, err)
	}

	common.LogVerbose("Exported name-description data to JSON: %s", jsonPath)
	return nil
}

// ExportToJSONForStore exports objects carrying the keyed layout metadata,
// without depending on objectsFilePatternPaths.
func ExportToJSONForStore(objects components.IList[datastore.IGlobalLocalizedTextObject], layout FileLayout, key string) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}
	return createJSONForStore(buildJSONEntries(objects), layout, key)
}

func createJSONForStore(dataObjectEntries []*JSONEntry, layout FileLayout, key string) error {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}
	fileName := strings.ReplaceAll(strings.ReplaceAll(key, "/", "_"), ".bin", ".json")
	jsonPath := filepath.Join(editsPath, fileName)

	stringsBytes, err := json.Marshal(dataObjectEntries)
	if err != nil {
		return fmt.Errorf("error marshaling data for %s: %w", jsonPath, err)
	}

	export := models.ObjectsFileExport{
		Metadata: models.NewObjectFileMetadataKeyed(layout.Version, layout.DirPattern, layout.FileName, key),
		Strings:  stringsBytes,
	}

	if err := models.SaveDataFile(export, jsonPath); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", jsonPath, err)
	}

	common.LogVerbose("Exported name-description data to JSON: %s", jsonPath)
	return nil
}

// getLocalizationKeys returns all available localization keys
func getLocalizationKeys() []string {
	var keys []string
	for key := range common.SupportedLanguages {
		keys = append(keys, key)
	}
	return keys
}

// getSortedLocalizationKeys returns localization keys sorted alphabetically
func getSortedLocalizationKeys() []string {
	localizationKeys := getLocalizationKeys()
	sort.Strings(localizationKeys)
	return localizationKeys
}

// buildEventStringDataJSON creates EventStringDataJSON from an event string with all localizations
func buildEventStringDataJSON(index int, str interface{ GetLocalizedString(string) string }, localizationKeys []string) EventStringDataJSON {
	stringData := EventStringDataJSON{
		Index: index,
		Text:  make(map[string]string),
	}

	for _, langKey := range localizationKeys {
		value := str.GetLocalizedString(langKey)
		stringData.Text[langKey] = value
	}

	return stringData
}

// currentGameVersion resolves the active game version for versioned event lookups.
func currentGameVersion() common.GameVersion {
	return interactions.CurrentGameVersion()
}

// processEventFromMemory processes an event from memory and creates EventFileDataJSON
func processEventFromMemory(eventID string, localizationKeys []string) *EventFileDataJSON {
	eventFile := event.GetEvent(currentGameVersion(), eventID)
	if eventFile == nil || eventFile.Strings == nil || len(eventFile.Strings) == 0 {
		return nil
	}

	eventData := EventFileDataJSON{
		ID:      eventFile.ID,
		Strings: make([]EventStringDataJSON, 0, len(eventFile.Strings)),
	}

	for i, str := range eventFile.Strings {
		stringData := buildEventStringDataJSON(i, str, localizationKeys)
		eventData.Strings = append(eventData.Strings, stringData)
	}

	if len(eventData.Strings) == 0 {
		common.LogVerbose("No strings found for event %s, skipping", eventID)
		return nil
	}

	return &eventData
}

// processEventFromFile processes an event from file and creates EventFileDataJSON
func processEventFromFile(eventID string, localizationKeys []string) *EventFileDataJSON {
	if common.IsVerboseMode() {
		common.LogVerbose("Exporting event file to JSON: %s", eventID)
	}

	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	eventFileStrings, err := event.ReadLocalizedEventStrings(eventID, version)
	if err != nil {
		common.LogVerbose("Error loading localized strings: %v", err)
		return nil
	}

	if len(eventFileStrings) == 0 {
		return nil
	}

	eventData := EventFileDataJSON{
		ID:      eventID,
		Strings: make([]EventStringDataJSON, 0, len(eventFileStrings)),
	}

	for i, str := range eventFileStrings {
		stringData := buildEventStringDataJSON(i, str, localizationKeys)
		eventData.Strings = append(eventData.Strings, stringData)
	}

	return &eventData
}

// writeEventJSONFile writes event data to a JSON file
func writeEventJSONFile(events []EventFileDataJSON, fileName string) error {
	if len(events) == 0 {
		common.LogVerbose("No events with string data found to export to JSON")
		return nil
	}

	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	filePath := filepath.Join(editsPath, common.WithVersionSuffix(fileName))

	export := make([]models.EventFileExport, 0, len(events))
	for _, e := range events {
		strings := make([]models.EventStringDataExport, 0, len(e.Strings))
		for _, s := range e.Strings {
			strings = append(strings, models.EventStringDataExport{Index: s.Index, Text: s.Text})
		}
		export = append(export, models.EventFileExport{
			Metadata: models.NewFileMetadata(models.NewFileInfoFromPath(models.EventBinaryPath(e.ID))),
			ID:       e.ID,
			Strings:  strings,
		})
	}

	if err := models.SaveDataFile(export, filePath); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", filePath, err)
	}

	common.LogVerbose("Exported event JSON file: %s", filePath)
	common.LogVerbose("Total events exported: %d", len(events))
	return nil
}

// ExportAllEventsToJSON exports all event data to a JSON file with all localizations.
func ExportAllEventsToJSON() error {
	fileName := "events_all_localizations.json"
	localizationKeys := getSortedLocalizationKeys()
	eventIDs := event.GetAllEventIDs(currentGameVersion())

	var allEvents []EventFileDataJSON
	var count int

	for _, eventID := range eventIDs {
		eventData := processEventFromMemory(eventID, localizationKeys)
		if eventData == nil {
			common.LogVerbose("Skipping event %s: no strings found", eventID)
			continue
		}

		allEvents = append(allEvents, *eventData)
		count++
	}

	common.LogVerbose("Total events processed: %d from %d", count, len(eventIDs))
	return writeEventJSONFile(allEvents, fileName)
}

// ExportEventsForLocalizationToJSON exports event data for a specific language to a JSON file.
func ExportEventsForLocalizationToJSON(languageCode string) error {
	eventIDs := event.GetAllEventIDs(currentGameVersion())
	localizationKeys := []string{languageCode}

	var allEvents []EventFileDataJSON

	for _, eventID := range eventIDs {
		eventData := processEventFromMemory(eventID, localizationKeys)
		if eventData == nil {
			continue
		}

		allEvents = append(allEvents, *eventData)
	}

	if len(allEvents) == 0 {
		return fmt.Errorf("no events with string data found for localization %s", languageCode)
	}

	fileName := fmt.Sprintf("events_%s.json", languageCode)
	return writeEventJSONFile(allEvents, fileName)
}

// ExportSingleEventToJSON exports a single event's data to a JSON file with all localizations.
func ExportSingleEventToJSON(eventId string) error {
	localizationKeys := getSortedLocalizationKeys()

	eventData := processEventFromFile(eventId, localizationKeys)
	if eventData == nil {
		return fmt.Errorf("no data found for event %s", eventId)
	}

	allEvents := []EventFileDataJSON{*eventData}
	fileName := "event_" + eventId + "_all_localizations.json"

	return writeEventJSONFile(allEvents, fileName)
}
