// Export agrupado de textos: UMA operação com um step por kind — o
// frontend chama UM binding (ExportArtifactBatch) e o modal de progresso
// é dirigido por eventos (backend/progress). Os formatos (json, strings,
// csv, xlsx) compartilham a MESMA extração; só o write muda.
package services

import (
	"fmt"
	"path/filepath"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters"
	_ "ffxresources/backend/formatters/csv"
	jsonfmt "ffxresources/backend/formatters/json"
	strfmt "ffxresources/backend/formatters/strings"
	_ "ffxresources/backend/formatters/xlsx"
	"ffxresources/backend/progress"
)

// ExportPlanItem é UMA ação do lote: um kind e os ids dele (vazio = tudo
// do kind; com seleção, os ids marcados na árvore).
type ExportPlanItem struct {
	Kind string   `json:"kind"`
	IDs  []string `json:"ids"`
}

// kindStepLabels são os rótulos PT dos steps do modal de export — os
// mesmos nomes da árvore (KIND_LABELS do frontend).
var kindStepLabels = map[string]string{
	KindEvents:     "Eventos",
	KindObjects:    "Sistema",
	KindMacro:      "Dicionário",
	KindLockit:     "Loc Kit",
	KindHelp:       "Painel de ajuda",
	KindBattleText: "Textos de batalha",
	KindCloud:      "Cloud",
	KindTutorial:   "Tutorial",
	KindMenuMain:   "Texto menu principal",
}

// ExportArtifactBatch exporta o lote como UMA operação agrupada: um step
// por kind, cada um com a própria barra (extração tick a tick — p.ex.
// "Eventos 0→301/301"). Formato: "json" | "strings" | "csv" | "xlsx".
// O progresso sai pelos eventos do wails (ProgressStart/Update/End) — o
// frontend só renderiza o modal; erro após o início do batch aparece lá.
func (s *MetadataService) ExportArtifactBatch(items []ExportPlanItem, version common.GameVersion, langs []string, format string) (paths []string, retErr error) {
	// Validação ANTES do batch: erros daqui ainda são toast do frontend.
	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case "json", "strings", "csv", "xlsx":
	default:
		return nil, fmt.Errorf("formato de export desconhecido: %s", format)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("nada para exportar: lote vazio")
	}
	for _, it := range items {
		if kind := strings.ToLower(strings.TrimSpace(it.Kind)); kindStepLabels[kind] == "" {
			return nil, fmt.Errorf("unknown kind: %s", it.Kind)
		}
	}

	labels := make([]string, len(items))
	for i, it := range items {
		labels[i] = kindStepLabels[strings.ToLower(strings.TrimSpace(it.Kind))]
	}
	batch := s.progress.NewBatch(fmt.Sprintf("Exportando textos (%s)…", exportFormatLabel(format)))
	steps := batch.Steps(labels...)
	// Sem emitter (testes/CLI): batch/Steps nil — cada ação ganha um step
	// nil, e TODO método de *Step é nil-safe (o Reporter vira no-op).
	if steps == nil {
		steps = make([]*progress.Step, len(items))
	}

	var all []string
	for i, it := range items {
		kind := strings.ToLower(strings.TrimSpace(it.Kind))
		ps, err := s.exportKind(steps[i], kind, format, version, it.IDs, langs)
		all = append(all, ps...)
		if err != nil {
			steps[i].Fail(err.Error())
			batch.End()
			return all, err
		}
	}
	batch.SetSummary(len(all), exportSummaryDestination(all))
	batch.End()
	return all, nil
}

// exportFormatLabel é o nome do formato nos títulos.
func exportFormatLabel(format string) string {
	switch strings.ToUpper(strings.TrimSpace(format)) {
	case "JSON":
		return "JSON"
	case "STRINGS":
		return "Strings"
	case "CSV":
		return "CSV"
	case "XLSX":
		return "XLSX"
	}
	return format
}

// exportSummaryDestination resume o destino para o estado final do modal:
// caminho único, ou "N arquivos em <diretório>".
func exportSummaryDestination(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	if len(paths) == 1 {
		return paths[0]
	}
	dir := filepath.Dir(paths[0])
	for _, p := range paths[1:] {
		if filepath.Dir(p) != dir {
			return fmt.Sprintf("%d arquivos", len(paths))
		}
	}
	return fmt.Sprintf("%d arquivos em %s", len(paths), dir)
}

// exportKind extrai e escreve o artefato de UM kind no formato pedido.
// O Reporter r é o step do batch: a extração (GetCollection → builders)
// faz SetTotal + um tick por entrada; o write em si é rápido e não ticka.
func (s *MetadataService) exportKind(r progress.Reporter, kind, format string, version common.GameVersion, ids, langs []string) (paths []string, retErr error) {
	r = progress.OrNop(r)
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind == KindEvents || kind == KindMacro {
		ids = s.normalizeIDs(ids)
	}
	r.Begin("", 0) // total indeterminado: os builders fazem SetTotal
	defer r.End()

	c, err := s.GetCollection(kind, version, helpExportIDs(kind, ids), r)
	if err != nil {
		return nil, err
	}
	c = filterExportRows(c)

	// Escreve pelo formato pedido — mesmos caminhos canônicos por escopo
	// (só a extensão muda), regra do dedup por arquivo preservada.
	switch format {
	case "csv", "xlsx":
		f, ok := formatters.Get(format)
		if !ok {
			return nil, fmt.Errorf("formatter %s não registrado", format)
		}
		ext := f.Extension()
		swapExt := func(jsonPath string) string {
			return strings.TrimSuffix(jsonPath, filepath.Ext(jsonPath)) + ext
		}
		write := func(col dto.Collection, path string) (string, error) {
			raw, werr := f.MarshalLangs(col, langs)
			if werr != nil {
				return "", werr
			}
			if werr = common.WriteBytesToFile(path, raw); werr != nil {
				return "", werr
			}
			return path, nil
		}
		switch kind {
		case KindEvents:
			path, perr := eventsExportPath(version, ids)
			if perr != nil {
				return nil, perr
			}
			p, werr := write(c, swapExt(path))
			return []string{p}, werr
		case KindMacro:
			path, perr := macroExportPath(version, ids)
			if perr != nil {
				return nil, perr
			}
			p, werr := write(c, swapExt(path))
			return []string{p}, werr
		case KindHelp:
			path, perr := jsonfmt.HelpJSONPath(version)
			if perr != nil {
				return nil, perr
			}
			p, werr := write(c, swapExt(path))
			return []string{p}, werr
		case KindObjects, KindBattleText, KindCloud, KindTutorial, KindMenuMain:
			if len(ids) == 0 {
				// Export completo: UM arquivo por kind (bulk, dedup global).
				var path string
				var perr error
				if kind == KindObjects {
					path, perr = objectsBulkPath(version)
				} else {
					path, perr = tableBulkPath(kind, version)
				}
				if perr != nil {
					return nil, perr
				}
				p, werr := write(c, swapExt(path))
				return []string{p}, werr
			}
			fallthrough
		case KindLockit:
			// Um arquivo por entrada (mesmo nome do JSON por entrada).
			for _, key := range c.SortedKeys() {
				jp, perr := jsonfmt.ObjectsJSONPath(key, version)
				if perr != nil {
					return nil, perr
				}
				p, werr := write(dto.Collection{key: c[key]}, swapExt(jp))
				if werr != nil {
					return nil, werr
				}
				paths = append(paths, p)
			}
			return paths, nil
		default:
			return nil, fmt.Errorf("unknown kind: %s", kind)
		}
	case "strings":
		f := strfmt.NewStringsFormatter()
		switch kind {
		case KindEvents:
			path, perr := eventsExportPath(version, ids)
			if perr != nil {
				return nil, perr
			}
			p, werr := f.WriteEventsFile(c, strfmt.StringsPathFor(path), langs)
			if werr != nil {
				return nil, werr
			}
			return []string{p}, nil
		case KindObjects:
			if len(ids) == 0 {
				path, perr := objectsBulkPath(version)
				if perr != nil {
					return nil, perr
				}
				p, werr := f.WriteObjectsFile(c, strfmt.StringsPathFor(path), langs)
				if werr != nil {
					return nil, werr
				}
				return []string{p}, nil
			}
			return f.WriteObjects(c, version, langs)
		case KindMacro:
			path, perr := macroExportPath(version, ids)
			if perr != nil {
				return nil, perr
			}
			p, werr := f.WriteMacroFile(c, path, langs)
			if werr != nil {
				return nil, werr
			}
			return []string{p}, nil
		case KindLockit:
			return f.WriteObjects(c, version, langs)
		case KindHelp:
			paths, werr := f.WriteHelp(c, version, langs)
			if werr != nil {
				return nil, werr
			}
			return paths, nil
		case KindBattleText, KindCloud, KindTutorial, KindMenuMain:
			if len(ids) == 0 {
				path, perr := tableBulkPath(kind, version)
				if perr != nil {
					return nil, perr
				}
				p, werr := f.WriteObjectsFile(c, strfmt.StringsPathFor(path), langs)
				if werr != nil {
					return nil, werr
				}
				return []string{p}, nil
			}
			return f.WriteObjects(c, version, langs)
		default:
			return nil, fmt.Errorf("unknown kind: %s", kind)
		}
	default: // json
		switch kind {
		case KindEvents:
			path, perr := eventsExportPath(version, ids)
			if perr != nil {
				return nil, perr
			}
			p, jerr := jsonfmt.NewJSONEventsFormatter().WriteEventsFile(c, path, langs)
			if jerr != nil {
				return nil, jerr
			}
			return []string{p}, nil
		case KindObjects:
			if len(ids) == 0 {
				// Export completo: UM arquivo com todas as entradas (cada
				// uma com sua metadata.key) e dedup GLOBAL entre elas.
				path, perr := objectsBulkPath(version)
				if perr != nil {
					return nil, perr
				}
				p, werr := jsonfmt.NewJSONObjectFormatter().WriteObjectsFile(c, path, langs)
				if werr != nil {
					return nil, werr
				}
				return []string{p}, nil
			}
			// Seleção: um arquivo por entrada (dedup intra-arquivo).
			return jsonfmt.NewJSONObjectFormatter().WriteObjects(c, version, langs)
		case KindMacro:
			path, perr := macroExportPath(version, ids)
			if perr != nil {
				return nil, perr
			}
			p, jerr := jsonfmt.NewJSONMacroFormatter().WriteMacroFile(c, path, langs)
			if jerr != nil {
				return nil, jerr
			}
			return []string{p}, nil
		case KindLockit:
			return jsonfmt.NewJSONObjectFormatter().WriteObjects(c, version, langs)
		case KindHelp:
			paths, jerr := jsonfmt.NewJSONHelpFormatter().WriteHelp(c, version, langs)
			if jerr != nil {
				return nil, jerr
			}
			return paths, nil
		case KindBattleText, KindCloud, KindTutorial, KindMenuMain:
			if len(ids) == 0 {
				path, perr := tableBulkPath(kind, version)
				if perr != nil {
					return nil, perr
				}
				p, werr := jsonfmt.NewJSONObjectFormatter().WriteObjectsFile(c, path, langs)
				if werr != nil {
					return nil, werr
				}
				return []string{p}, nil
			}
			return jsonfmt.NewJSONObjectFormatter().WriteObjects(c, version, langs)
		default:
			return nil, fmt.Errorf("unknown kind: %s", kind)
		}
	}
}
