// Export CSV/XLSX do FRONTEND: mesmas regras do JSON/.strings — artefato em
// mods/edits com o nome canônico por escopo (só a extensão muda).
package services

import (
	"fmt"
	"path/filepath"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/core/progress"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters"
	_ "ffxresources/backend/formatters/csv"
	jsonfmt "ffxresources/backend/formatters/json"
	_ "ffxresources/backend/formatters/xlsx"
)

// ExportArtifact exporta o lote no formato pedido ("json"|"strings"|"csv"|"xlsx").
func (s *MetadataService) ExportArtifact(kind string, version common.GameVersion, ids, langs []string, format string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return s.ExportJSON(kind, version, ids, langs)
	case "strings":
		return s.ExportStrings(kind, version, ids, langs)
	case "csv", "xlsx":
		return s.exportRegistry(kind, version, ids, langs, strings.ToLower(strings.TrimSpace(format)))
	default:
		return nil, fmt.Errorf("formato de export desconhecido: %s", format)
	}
}

// exportRegistry escreve o artefato csv/xlsx no caminho canônico do
// frontend: o MESMO do JSON do escopo (events/bulk/per-entry), extensão trocada.
func (s *MetadataService) exportRegistry(kind string, version common.GameVersion, ids, langs []string, format string) (paths []string, retErr error) {
	f, ok := formatters.Get(format)
	if !ok {
		return nil, fmt.Errorf("formatter %s não registrado", format)
	}
	ext := f.Extension()

	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind == KindEvents || kind == KindMacro {
		ids = s.normalizeIDs(ids)
	}
	progress.Begin(fmt.Sprintf("Extraindo texto (%s, %s)", kind, format), len(ids))
	defer func() {
		if retErr != nil {
			progress.Issue("", retErr.Error())
		}
		progress.End()
	}()

	c, err := s.GetCollection(kind, version, helpExportIDs(kind, ids))
	if err != nil {
		return nil, err
	}
	c = filterExportRows(c)
	progress.SetTotal(len(c))

	// swapExt troca a extensão .json do caminho canônico pelo formato pedido.
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
}
