package event

import (
	"fmt"
	"path/filepath"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/progress"
	"ffxresources/backend/datastore"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
)

// EventsBinaryFile carrega/salva os eventos de localization no binário.
// Este pacote desconhece JSON/DTO: exportação de texto vive em
// backend/builders (monta dto.Collection) + backend/formatters/json
// (serializa e escreve o arquivo). Em SaveToBinary o parâmetro string
// é usado como: "" = tudo, ou EventID.
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

// eventsFolder resolve o diretório de eventos para uma localização.
// Descoberta na árvore ORIGINAL: com mods habilitado, a pasta
// mods/<loc>/event/obj_ps3 existe após o primeiro import com apenas alguns
// eventos — enumerar nela carregaria só o que está em mods. O fluxo correto:
// enumerar os eventos em gamefiles e, por arquivo, preferir o binário de mods
// (ReadLocalizedStringFiles via NewFileAccessor). Fallback: sem árvore
// original, usa o caminho resolvido normal (mods).
func (b *EventsBinaryFile) eventsFolder(localization string) (common.FileAccessor, error) {
	version := b.Version
	if version.String() == "" {
		version = common.GameVersionFFX
	}
	folder, err := common.NewRealFileAccessor(filepath.Join(
		common.GetPathRootForVersion(version),
		"new_"+localization+"pc",
		"event",
		"obj_ps3",
	))
	if err != nil {
		return common.FileAccessor{}, err
	}
	if folder.Exists {
		return folder, nil
	}
	return common.NewFileAccessor(filepath.Join(
		common.GetLocalizationRootForVersion(version, localization),
		"event",
		"obj_ps3",
	))
}

func (b *EventsBinaryFile) LoadFromBinary() error {
	eventsFolder, err := b.eventsFolder(common.DefaultLocalization)
	if err != nil {
		return fmt.Errorf("failed to resolve events directory: %w", err)
	}
	if b.Version.String() == "" {
		b.Version = currentVersionOrDefault()
	}
	if err := ffxencoding.EnsureAllCharsetsLoaded(b.Version); err != nil {
		return fmt.Errorf("charset maps not loaded: %w", err)
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

	// Sem barra de progresso: a carga é implícita (boot/carga de árvore),
	// à frente dela já existem os indicadores próprios da árvore — um modal
	// aqui aparecia DOIS no boot (a pré-carga das outras versões dispara
	// outra carga em background). Falhas por evento são PULADAS com aviso
	// individual (Issue → toast com cap + arquivo de diagnóstico) e a carga
	// segue com o que deu certo.

	for _, info := range loadedInfos.Items() {
		eventFile, err := ReadCompleteEventFile(info)
		if err != nil {
			progress.Issue(info.EventID, fmt.Sprintf("evento pulado (falha de leitura): %s: %v", info.EventID, err))
			common.LogWarning("failed to read event file %s: %v", info.EventID, err)
			continue
		}
		if eventFile == nil {
			progress.Issue(info.EventID, fmt.Sprintf("evento pulado (sem conteúdo): %s", info.EventID))
			continue
		}
		if b.Objects == nil {
			b.Objects = components.NewEmptyList[datastore.IGlobalLocalizedTextObject]()
		}
		b.Objects.Add(NewEventKeyedStringFile(info, eventFile.Strings))
	}

	b.Infos = loadedInfos
	// Vigia de frescura: carimbo físico de cada evento carregado — binário
	// modificado no disco (tradução copiada em mods/) recarrega o evento
	// na próxima leitura (EnsureEventFresh).
	ids := make([]string, 0, loadedInfos.Len())
	for _, info := range loadedInfos.Items() {
		ids = append(ids, info.EventID)
	}
	stampAllEvents(b.Version, ids)
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
