package event

import (
	"fmt"
	"path/filepath"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
)

// EventsBinaryFile carrega/exporta/importa/salva os eventos de localization.
// Implementa datastore.IBinaryFile. Em ExportToJson, ImportFromJson e
// SaveToBinary o parâmetro string é usado como: "" = tudo, ou EventID.
type EventsBinaryFile struct {
	Version common.GameVersion
	Infos   components.IList[*models.EventFileInfo]
	Objects components.IList[datastore.IGlobalLocalizedTextObject]
}

func NewEventsBinaryFile(version common.GameVersion) *EventsBinaryFile {
	return &EventsBinaryFile{
		Version: version,
		Objects: components.NewEmptyList[datastore.IGlobalLocalizedTextObject](),
	}
}

func currentVersionOrDefault() common.GameVersion {
	if svc := interactions.NewInteractionService(); svc != nil {
		return svc.FFXAppConfig().GetGameVersion()
	}
	return common.GameVersionFFX
}

func (b *EventsBinaryFile) eventsFolder(localization string) (common.FileAccessor, error) {
	return common.NewFileAccessor(filepath.Join(common.GetLocalizationRootForVersion(b.Version, localization), "event", "obj_ps3"))
}

func (b *EventsBinaryFile) LoadFromBinary() error {
	eventsFolder, err := b.eventsFolder(common.DefaultLocalization)
	if err != nil {
		return fmt.Errorf("failed to resolve events directory: %w", err)
	}
	if b.Version.String() == "" {
		b.Version = currentVersionOrDefault()
	}

	first := true

	loadedInfos := components.NewList[*models.EventFileInfo](0)
	for _, localization := range SortedSupportedLocalizations() {
		eventInfos, err := discoverEventFiles(eventsFolder, localization, b.Version)
		if err != nil {
			if first {
				return fmt.Errorf("failed to discover event files: %w", err)
			}
			common.LogVerbose("failed to discover event files for %s: %v", localization, err)
			continue
		}

		for _, info := range eventInfos {
			if first {
				loadedInfos.Add(info)
				continue
			}
			if loadedInfos.TryAdd(info) {
				common.LogInfo("event id added from %s: %s", localization, info.EventID)
			}
		}
		first = false
	}

	for _, info := range loadedInfos.Items() {
		eventFile, err := ReadCompleteEventFile(info)
		if err != nil {
			common.LogVerbose("failed to read event file %s: %v", info.EventID, err)
			continue
		}
		if eventFile == nil {
			continue
		}
		if b.Objects == nil {
			b.Objects = components.NewEmptyList[datastore.IGlobalLocalizedTextObject]()
		}
		b.Objects.Add(NewEventKeyedStringFile(info, eventFile.Strings))
	}

	b.Infos = loadedInfos
	common.LogInfo("events loaded: %d", loadedInfos.Len())
	return nil
}

func (b *EventsBinaryFile) eventIDs(filePath string) ([]string, error) {
	if strings.TrimSpace(filePath) == "" {
		return GetAllEventIDs(b.Version), nil
	}
	eventID := eventIDFromParam(filePath)
	if eventID == "" {
		return nil, fmt.Errorf("invalid event selector: %s", filePath)
	}
	if GetEvent(b.Version, eventID) == nil {
		return nil, fmt.Errorf("event not found: %s", eventID)
	}
	return []string{eventID}, nil
}

func eventIDFromParam(filePath string) string {
	base := filepath.Base(strings.TrimSpace(filePath))
	base = strings.TrimSuffix(base, filepath.Ext(base))
	parts := strings.Split(base, "_")
	if len(parts) > 1 {
		last := parts[len(parts)-1]
		if len(last) >= 2 {
			return last
		}
	}
	if len(base) >= 2 {
		return base
	}
	return ""
}

func (b *EventsBinaryFile) ExportToJson(filePath string) error {
	eventIDs, err := b.eventIDs(filePath)
	if err != nil {
		return err
	}
	if len(eventIDs) == 0 {
		return fmt.Errorf("no events loaded")
	}
	if len(eventIDs) == 1 {
		return exportSingleEventToJSON(b.Version, eventIDs[0])
	}
	return ExportAllEventsToJSONForVersion(b.Version)
}

func (b *EventsBinaryFile) ImportFromJson(filePath string) error {
	eventIDs, err := b.eventIDs(filePath)
	if err != nil {
		return err
	}
	for _, eventID := range eventIDs {
		if err := ImportEventDataFromJsonFile(b.Version, eventID); err != nil {
			return err
		}
	}
	return nil
}

func (b *EventsBinaryFile) SaveToBinary(filePath string) error {
	eventIDs, err := b.eventIDs(filePath)
	if err != nil {
		return err
	}
	for _, eventID := range eventIDs {
		if err := ExportEventStringsToLocalizations(b.Version, eventID); err != nil {
			return err
		}
	}
	return nil
}

func (b *EventsBinaryFile) GetObjects() components.IList[datastore.IGlobalLocalizedTextObject] {
	if b.Objects == nil {
		b.Objects = components.NewEmptyList[datastore.IGlobalLocalizedTextObject]()
	}
	return b.Objects
}
