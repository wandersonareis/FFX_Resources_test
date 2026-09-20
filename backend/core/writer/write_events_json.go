package writer

import (
	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/formatters/json"
	"ffxresources/backend/interactions"
	"fmt"
	"sort"
)

func currentGameVersion() common.GameVersion {
	return interactions.CurrentGameVersion()
}

func getSortedEventIDs() []string {
	eventIDs := event.GetAllEventIDs(currentGameVersion())
	sort.Strings(eventIDs)
	return eventIDs
}

// ExportEventsToJSON exporta os eventos indicados (ids vazio = tudo)
// pelo pipeline DTO: builders monta a Collection pronta e o formatter
// JSON serializa e escreve o arquivo. Fluxo estrito: ids pedidos e não
// encontrados são logados; se zero for encontrado, retorna erro sem
// cair para "tudo".
func ExportEventsToJSON(ids []string) error {
	version := currentGameVersion()
	collection, err := builders.BuildEventsDTO(version, ids)
	if err != nil {
		return err
	}
	if _, err := json.NewJSONEventsFormatter().WriteEvents(collection, version); err != nil {
		return err
	}
	return nil
}

// ExportAllLocalizationsToJSON writes event files as JSON for all localizations
// Creates JSON files with event strings for each language in the edits/ directory
//
// JSON Format (novo, quebra autorizada):
//   - Mapa raiz chaveado pelo nome do arquivo sem extensão (= eventID)
//   - Cada entrada contém metadata (mesmos campos do gerador legado) e rows
//   - Cada row tem index, hash xxHash64 por idioma e text por idioma
func ExportAllLocalizationsToJSON() {
	if err := ExportEventsToJSON(nil); err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	if common.IsVerboseMode() {
		fmt.Printf("Total events processed: %d\n", len(getSortedEventIDs()))
	}
}
