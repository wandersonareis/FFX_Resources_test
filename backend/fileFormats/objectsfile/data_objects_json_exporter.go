package objectsfile

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

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

type textGroup interface {
	exportTo(data *JSONEntry, locKey string)
}

type commandTexts struct {
	name                  datastore.IGlobalLocalizedKeyedStringObject
	simplifiedName        datastore.IGlobalLocalizedKeyedStringObject
	description           datastore.IGlobalLocalizedKeyedStringObject
	simplifiedDescription datastore.IGlobalLocalizedKeyedStringObject
}

func resolveCommandTexts(obj datastore.IGlobalLocalizedTextObject) (commandTexts, bool) {
	c := commandTexts{
		name:                  obj.GetKeyedString("name"),
		simplifiedName:        obj.GetKeyedString("simplifiedName"),
		description:           obj.GetKeyedString("description"),
		simplifiedDescription: obj.GetKeyedString("simplifiedDescription"),
	}
	return c, c.name != nil || c.simplifiedName != nil ||
		c.description != nil || c.simplifiedDescription != nil
}

func (c commandTexts) exportTo(data *JSONEntry, locKey string) {
	data.Name = fillLocalized(data.Name, c.name, locKey)
	data.SimplifiedName = fillLocalized(data.SimplifiedName, c.simplifiedName, locKey)
	data.Description = fillLocalized(data.Description, c.description, locKey)
	data.SimplifiedDescription = fillLocalized(data.SimplifiedDescription, c.simplifiedDescription, locKey)
}

type effectTexts struct {
	effect datastore.IGlobalLocalizedKeyedStringObject
}

func resolveEffectTexts(obj datastore.IGlobalLocalizedTextObject) (effectTexts, bool) {
	e := effectTexts{effect: obj.GetKeyedString("effect")}
	return e, e.effect != nil
}

func (e effectTexts) exportTo(data *JSONEntry, locKey string) {
	data.Effect = fillLocalized(data.Effect, e.effect, locKey)
}

type abilityTexts struct {
	abilities []datastore.IGlobalLocalizedKeyedStringObject
}

func resolveAbilityTexts(obj datastore.IGlobalLocalizedTextObject) (abilityTexts, bool) {
	var a abilityTexts
	for idx := 1; ; idx++ {
		seg := obj.GetKeyedString(fmt.Sprintf("ability%d", idx))
		if seg == nil {
			break
		}
		a.abilities = append(a.abilities, seg)
	}
	return a, len(a.abilities) > 0
}

func (a abilityTexts) exportTo(data *JSONEntry, locKey string) {
	if data.Abilities == nil {
		data.Abilities = make([]map[string]string, len(a.abilities))
		for i := range data.Abilities {
			data.Abilities[i] = make(map[string]string)
		}
	}
	for idx, seg := range a.abilities {
		data.Abilities[idx] = fillLocalized(data.Abilities[idx], seg, locKey)
	}
}

type monsterTexts struct {
	sensor           datastore.IGlobalLocalizedKeyedStringObject
	simplifiedSensor datastore.IGlobalLocalizedKeyedStringObject
	scan             datastore.IGlobalLocalizedKeyedStringObject
	simplifiedScan   datastore.IGlobalLocalizedKeyedStringObject
}

func resolveMonsterTexts(obj datastore.IGlobalLocalizedTextObject) (monsterTexts, bool) {
	m := monsterTexts{
		sensor:           obj.GetKeyedString("sensorText"),
		simplifiedSensor: obj.GetKeyedString("simplifiedSensorText"),
		scan:             obj.GetKeyedString("scanText"),
		simplifiedScan:   obj.GetKeyedString("simplifiedScanText"),
	}
	return m, m.sensor != nil || m.simplifiedSensor != nil ||
		m.scan != nil || m.simplifiedScan != nil
}

func (m monsterTexts) exportTo(data *JSONEntry, locKey string) {
	data.SensorText = fillLocalized(data.SensorText, m.sensor, locKey)
	data.SimplifiedSensorText = fillLocalized(data.SimplifiedSensorText, m.simplifiedSensor, locKey)
	data.ScanText = fillLocalized(data.ScanText, m.scan, locKey)
	data.SimplifiedScanText = fillLocalized(data.SimplifiedScanText, m.simplifiedScan, locKey)
}

type lastMissionTexts struct {
	effect          datastore.IGlobalLocalizedKeyedStringObject
	effectDescription datastore.IGlobalLocalizedKeyedStringObject
}

func resolveLastMissionTexts(obj datastore.IGlobalLocalizedTextObject) (lastMissionTexts, bool) {
	l := lastMissionTexts{
		effect:          obj.GetKeyedString("effect"),
		effectDescription: obj.GetKeyedString("effectDescription"),
	}
	return l, l.effect != nil || l.effectDescription != nil
}

func (l lastMissionTexts) exportTo(data *JSONEntry, locKey string) {
	data.Effect = fillLocalized(data.Effect, l.effect, locKey)
	data.EffectDescription = fillLocalized(data.EffectDescription, l.effectDescription, locKey)
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

func resolveTextGroups(obj datastore.IGlobalLocalizedTextObject) []textGroup {
	var groups []textGroup
	if g, ok := resolveCommandTexts(obj); ok {
		groups = append(groups, g)
	}
	if g, ok := resolveEffectTexts(obj); ok {
		groups = append(groups, g)
	}
	if g, ok := resolveAbilityTexts(obj); ok {
		groups = append(groups, g)
	}
	if g, ok := resolveMonsterTexts(obj); ok {
		groups = append(groups, g)
	}
	if g, ok := resolveLastMissionTexts(obj); ok {
		groups = append(groups, g)
	}
	if g, ok := resolveWeaponsTexts(obj); ok {
		groups = append(groups, g)
	}
	return groups
}

func (e *JSONEntry) hasContent() bool {
	if len(e.Name) > 0 || len(e.SimplifiedName) > 0 ||
		len(e.Description) > 0 || len(e.SimplifiedDescription) > 0 ||
		len(e.Effect) > 0 || len(e.EffectDescription) > 0 ||
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

func ExportToJSON(objects components.IList[datastore.IGlobalLocalizedTextObject], jsonFileName string) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}

	var jsonEntries []*JSONEntry

	objects.RangeIndex(func(i int, obj datastore.IGlobalLocalizedTextObject) {
		if obj == nil {
			common.LogVerbose("Object %d is nil, skipping", i)
			return
		}

		groups := resolveTextGroups(obj)
		if len(groups) == 0 {
			return
		}

		data := &JSONEntry{ID: i}

		for locKey := range common.SupportedLanguages {
			for _, g := range groups {
				g.exportTo(data, locKey)
			}
		}

		if data.hasContent() {
			jsonEntries = append(jsonEntries, data)
		}
	})

	return createJSON(jsonEntries, jsonFileName)
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

// objectsFilePatternPaths maps each exported objectsfile JSON to the pattern path of
// the original game binary it was extracted from, so its metadata can be reconstructed.
var objectsFilePatternPaths = map[string]string{
	"key_items_all_localizations.json":       "battle/kernel/important.bin",
	"commands_all_localizations.json":        "battle/kernel/command.bin",
	"items_all_localizations.json":           "battle/kernel/item.bin",
	"arms_all_localizations.json":            "battle/kernel/arms_txt.bin",
	"config_text_all_localizations.json":     "battle/kernel/config_txt.bin",
	"item_commands_all_localizations.json":   "battle/kernel/item_txt.bin",
	"main_menu_all_localizations.json":       "battle/kernel/mmain_txt.bin",
	"player_room_all_localizations.json":     "battle/kernel/ply_rom.bin",
	"battle_text_all_localizations.json":     "battle/kernel/btl_txt.bin",
	"battle_end_text_all_localizations.json": "battle/kernel/btlend_txt.bin",
	"monster_magic1_all_localizations.json":  "battle/kernel/monmagic1.bin",
	"monster_magic2_all_localizations.json":  "battle/kernel/monmagic2.bin",
	"build_all_localizations.json":           "battle/kernel/build_txt.bin",
	"names_all_localizations.json":           "battle/kernel/name_txt.bin",

	// FFX-2 (v2) kernel exclusivos.
	"a_ability_all_localizations.json":     "battle/kernel/a_ability.bin",
	"accessory_all_localizations.json":     "battle/kernel/accessory.bin",
	"job_all_localizations.json":           "battle/kernel/job.bin",
	"menu_text_all_localizations.json":     "battle/kernel/menu_txt.bin",
	"monster_magic_all_localizations.json": "battle/kernel/monmagic.bin",
	"monster_all_localizations.json":       "battle/kernel/monster.bin",
	"monster2_all_localizations.json":      "battle/kernel/monster2.bin",
	"oversoul_all_localizations.json":      "battle/kernel/oversoul.bin",
	"plate_all_localizations.json":         "battle/kernel/plate.bin",
	"player_save_all_localizations.json":   "battle/kernel/ply_save.bin",
	"save_text_all_localizations.json":     "battle/kernel/save_txt.bin",
}

// binaryMetadataForFile returns metadata describing the original binary for a given
// objectsfile export filename, or nil when the mapping is unknown.
func binaryMetadataForFile(fileName string) *models.FileMetadata {
	pattern, ok := objectsFilePatternPaths[fileName]
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
