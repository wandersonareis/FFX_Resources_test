// Scanner one-off: conta as ocorrências das tags nomeadas (PC, MCR, BUTTON,
// ICON) no texto extraído de events, objects e macros (idioma us) e grava
// mods/edits/tag_frequency_<versão>.json — o ranking de uso consumido pelo
// catálogo do editor (binding GetTagCatalog).
//
// O ranking só REORDENA as sugestões; valores fora do arquivo mantêm a ordem
// fixa por índice, e o arquivo nunca exclui nada do catálogo.
//
// Uso: go run ./scripts/tagfreq FFX|FFX-2 [raizDoJogo]
//
//	raiz padrão: <repositório>/build/bin/data
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/interactions"
)

// Identidade da tag sem o texto humano — mesma chave usada pelo catálogo:
// {PC:01:YUNA} → "PC:01", {MCR:s06:l3F:"…"} → "MCR:s06:l3F".
var reNamedTag = regexp.MustCompile(`\{(PC|BUTTON|ICON):([0-9A-Fa-f]{2}):|\{MCR:(s[0-9A-Fa-f]{2}:l[0-9A-Fa-f]{2}):`)

func main() {
	if len(os.Args) < 2 {
		panic("uso: go run ./scripts/tagfreq FFX|FFX-2 [raizDoJogo]")
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

	root := filepath.Join("build", "bin", "data")
	if len(os.Args) > 2 {
		root = os.Args[2]
	}
	common.SetGameFilesRoot(root)
	common.SetCurrentGameVersion(version)

	// Config apontando para a raiz escolhida: os loaders de objects consultam
	// o config (GameFilesLocation) e NewInteractionService() padrão
	// sobrescreveria o GameFilesRoot com o diretório do executável.
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
	total := 0
	record := func(text string) {
		for _, m := range reNamedTag.FindAllStringSubmatch(text, -1) {
			var identity string
			if m[1] != "" {
				identity = m[1] + ":" + m[2]
			} else {
				identity = "MCR:" + m[3]
			}
			counts[identity]++
			total++
		}
	}
	// Events (todas as entradas) — idioma us.
	ev := event.NewEventsBinaryFile(version)
	if err := ev.LoadFromBinary(); err != nil {
		panic(err)
	}
	objs := ev.GetObjects().Items()
	for _, obj := range objs {
		record(obj.ToString(common.DefaultLocalization))
	}

	// Objects (battle/menu/help etc.) via FileLayouts da versão.
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
		for _, obj := range list.Items() {
			record(obj.ToString(common.DefaultLocalization))
		}
	}

	// Macros: os próprios textos do dicionário (podem conter tags).
	if macros := datastore.GetMacros(version); macros != nil {
		macros.ForEach(func(_ int, m datastore.IGlobalLocalizedMacroStringObject) {
			if m != nil {
				record(m.GetLocalizedString(common.DefaultLocalization))
			}
		})
	}

	if len(counts) == 0 {
		fmt.Println("nenhuma tag nomeada encontrada no texto extraído")
		return
	}

	outDir := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}
	outPath := filepath.Join(outDir, common.WithVersionSuffixFor("tag_frequency.json", version))
	raw, err := json.MarshalIndent(counts, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(outPath, append(raw, '\n'), 0o644); err != nil {
		panic(err)
	}

	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return counts[keys[i]] > counts[keys[j]] })
	fmt.Printf("total de tags nomeadas: %d (identidades: %d)\n", total, len(keys))
	for _, key := range keys {
		fmt.Printf("  {%s} × %d\n", key, counts[key])
	}
	fmt.Println("gravado em:", outPath)
}
