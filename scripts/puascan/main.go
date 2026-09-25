// Scanner one-off: decodifica todos os events (todas as localizações) e as
// macros de um jogo do testData e agrega as ocorrências de tokens {PUA:XX:...}
// emitidos pelo decode de slots duplicados da tabela de encoding.
//
// Uso: go run ./scripts/puascan FFX | FFX-2
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/interactions"
)

var rePUA = regexp.MustCompile(`\{PUA:([0-9A-Fa-f]+):`)

func main() {
	if len(os.Args) < 2 {
		panic("uso: go run ./scripts/puascan FFX|FFX-2")
	}

	var version common.GameVersion
	switch os.Args[1] {
	case "FFX":
		version = common.GameVersionFFX
	case "FFX-2":
		version = common.GameVersionFFX2
	default:
		panic("jogo desconhecido: " + os.Args[1])
	}

	common.SetGameFilesRoot(filepath.Join(`F:\ffxWails\FFX_Resources\testData`, os.Args[1], "binary"))
	common.SetCurrentGameVersion(version)

	// Cria o singleton de interactions com config apontando para o testData:
	// os loaders de objects consultam o config (GameFilesLocation) e o
	// NewInteractionService() padrão sobrescreveria o GameFilesRoot com o
	// diretório do executável.
	cfg := interactions.NewAppConfig()
	cfg.SetLocation("GameFilesLocation", common.GameFilesRoot)
	interactions.NewInteractionServiceWithConfig(cfg)

	for _, cs := range common.Charsets {
		if err := ffxencoding.PrepareCharset(version, cs); err != nil {
			panic(err)
		}
	}
	if err := reader.PrepareVersion(version); err != nil {
		panic(err)
	}

	counts := map[string]int{}
	example := map[string]string{}
	total := 0
	record := func(text, where string) {
		for _, m := range rePUA.FindAllStringSubmatch(text, -1) {
			code := strings.ToUpper(m[1])
			counts[code]++
			total++
			if example[code] == "" {
				i := strings.Index(text, m[0])
				end := i + len(m[0]) + 40
				if end > len(text) {
					end = len(text)
				}
				example[code] = fmt.Sprintf("%s | …%s…", where, text[max(0, i-30):end])
			}
		}
	}

	ev := event.NewEventsBinaryFile(version)
	if err := ev.LoadFromBinary(); err != nil {
		panic(err)
	}
	objs := ev.GetObjects().Items()
	fmt.Printf("eventos carregados: %d\n", len(objs))
	for _, obj := range objs {
		for loc := range common.SupportedLanguages {
			text := obj.ToString(loc)
			record(text, loc)
		}
	}

	if macros := datastore.GetMacros(version); macros != nil {
		macroTexts := 0
		macros.ForEach(func(_ int, m datastore.IGlobalLocalizedMacroStringObject) {
			for loc := range common.SupportedLanguages {
				if content, ok := m.GetLocalizedContent(loc); ok && content != nil && !content.IsEmpty() {
					macroTexts++
					record(content.GetString(), "macro:"+loc)
				}
			}
		})
		fmt.Printf("macros (pares idioma/texto não vazios): %d\n", macroTexts)
	}

	// Objects (battle/menu/help etc.) via FileLayouts da versão.
	objCount := 0
	for _, layout := range objectsfile.FileLayouts {
		if layout.Version != version {
			continue
		}
		binFile, err := objectsfile.LoadObjectFileFromStoreByLayout(layout)
		if err != nil || binFile == nil {
			continue
		}
		list := binFile.GetObjects()
		if list == nil || list.IsEmpty() {
			continue
		}
		objCount++
		for _, obj := range list.Items() {
			for loc := range common.SupportedLanguages {
				record(obj.ToString(loc), "object:"+layout.FileName)
			}
		}
	}
	fmt.Printf("object files com texto: %d\n", objCount)

	fmt.Printf("total de tokens {PUA:...} no texto extraído: %d\n", total)
	if len(counts) == 0 {
		fmt.Println("NENHUM slot duplicado apareceu no texto decodificado.")
		return
	}
	codes := make([]string, 0, len(counts))
	for code := range counts {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	for _, code := range codes {
		fmt.Printf("  {PUA:%s} × %d  ex.: %s\n", code, counts[code], example[code])
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
