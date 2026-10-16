// Codecs avulsos: decodificar um binário APONTADO (bytes crus, em qualquer
// pasta) no DTO e aplicar o DTO de volta em BYTES — sem árvore do jogo, sem
// store, sem mods/. Cada kind usa o mesmo motor do app (builders/appliers);
// só a fonte e o destino mudam.
//
// Mapa de dependências (ver TODO do runtime.go):
//
//	events/eventtable → encoding + macrodic (tags MCR), codec FieldString;
//	objects           → chunkmap (SchemaBytesLoaders/SchemaRebuilders);
//	macro             → container de 16 chunks (bytes → DTO → bytes);
//	lockit            → registros CRLF por idioma (só encoding);
//	help              → painel .sps2 (HelpBinaryFile/HelpKeyedStringFile).
package appcore

import (
	"fmt"
	"path/filepath"
	"strings"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/eventtable"
	"ffxresources/backend/fileFormats/helpfile"
	"ffxresources/backend/fileFormats/lockit"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/fileFormats/objectsfile/chunkmap"
	"ffxresources/backend/services"
)

// decodeAvulso decodifica o binário apontado no DTO pronto de sua entrada
// (chave da Collection = id canônico). A detecção já rodou: d.Kind diz qual
// codec usar; lang é o idioma do ARQUIVO (lockit/events por localização).
func decodeAvulso(d Detection, version common.GameVersion, path string, data []byte, lang string) (dto.Collection, error) {
	switch d.Kind {
	case services.KindEvents:
		objs, err := parseFieldStringObjects(data, version, lang)
		if err != nil {
			return nil, err
		}
		id := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		entry, ok := builders.BuildEventEntryDTOFrom(id, version, objs)
		if !ok {
			return nil, Incompatible(path, services.KindEvents, "", "binário sem textos")
		}
		return dto.Collection{id: entry}, nil

	case services.KindBattleText, services.KindCloud, services.KindTutorial, services.KindMenuMain:
		id := tableIDFor(d.Kind, path)
		if _, ok := eventtable.RelPath(d.Kind, id); !ok {
			return nil, Incompatible(path, d.Kind, "", "id de tabela desconhecido (fora da família "+d.Kind+")")
		}
		objs, err := parseFieldStringObjects(data, version, lang)
		if err != nil {
			return nil, err
		}
		entry, ok := builders.BuildTableEntryDTOFrom(d.Kind, id, version, objs)
		if !ok {
			return nil, Incompatible(path, d.Kind, "", "binário sem textos")
		}
		return dto.Collection{id: entry}, nil

	case services.KindObjects:
		_, load, _, ok := objectSchemaFor(path, version)
		if !ok {
			return nil, Incompatible(path, services.KindObjects, "", "binário de objects sem schema registrado (arquivo não servido pelo app)")
		}
		return load(version, data)

	case services.KindMacro:
		containers, err := macroContainerFromBytes(data, version, lang)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		return builders.BuildMacroDTOFromContainers(version, containers)

	case services.KindLockit:
		layout, ok := lockitLayoutFor(path, version)
		if !ok {
			return nil, Incompatible(path, services.KindLockit, "", "stem não é um kit de localização conhecido")
		}
		f, err := lockit.LoadFromBytes(layout, lang, data)
		if err != nil {
			return nil, err
		}
		return builders.BuildLockitDTOFrom([]*lockit.LockitFile{f})

	case services.KindHelp:
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		panel, err := helpPanelFromBytes(data, name, version, lang)
		if err != nil {
			return nil, err
		}
		entry, ok := builders.BuildHelpEntryDTOFrom(name, version, panel)
		if !ok {
			return nil, Incompatible(path, services.KindHelp, "", "painel sem segmentos de texto")
		}
		return dto.Collection{name: entry}, nil

	default:
		return nil, Incompatible(path, "", d.Format.String(), "kind sem codec avulso")
	}
}

// applyAvulso aplica as rows do artefato (uma entrada já filtrada pelo
// casamento de basename) nos bytes do binário apontado e devolve os bytes
// reconstruídos — nada é gravado aqui.
func applyAvulso(d Detection, version common.GameVersion, path string, data []byte, lang string, c dto.Collection) ([]byte, error) {
	switch d.Kind {
	case services.KindEvents:
		objs, err := parseFieldStringObjectsAllLangs(data, version, lang)
		if err != nil {
			return nil, err
		}
		return applyWithTableScope(services.KindEvents, c, objs, version, lang, data)

	case services.KindBattleText, services.KindCloud, services.KindTutorial, services.KindMenuMain:
		objs, err := parseFieldStringObjectsAllLangs(data, version, lang)
		if err != nil {
			return nil, err
		}
		return applyWithTableScope(d.Kind, c, objs, version, lang, data)

	case services.KindObjects:
		_, _, rebuild, ok := objectSchemaFor(path, version)
		if !ok {
			return nil, Incompatible(path, services.KindObjects, "", "binário de objects sem schema registrado")
		}
		entry, _, err := singleEntry(c)
		if err != nil {
			return nil, err
		}
		return rebuild(version, data, entry.Rows)

	case services.KindMacro:
		// Base = o próprio container apontado; o artefato sobrepõe os chunks
		// que traz. O rebuild reconstrói o dicionário INTEIRO (um arquivo só
		// por localização) e a saída é o container do idioma do arquivo.
		containers, err := macroContainerFromBytes(data, version, lang)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		base, err := builders.BuildMacroDTOFromContainers(version, containers)
		if err != nil {
			return nil, err
		}
		merged := make(dto.Collection, len(base)+len(c))
		for id, entry := range base {
			merged[id] = entry
		}
		for id, entry := range c {
			merged[id] = entry
		}
		rebuilt, err := builders.RebuildMacroContainers(version, builders.ResolveDedupRefs(merged))
		if err != nil {
			return nil, err
		}
		container, ok := rebuilt[lang]
		if !ok {
			return nil, fmt.Errorf("%s: rebuild não produziu o container %s", filepath.Base(path), lang)
		}
		return container.Bytes, nil

	case services.KindLockit:
		layout, ok := lockitLayoutFor(path, version)
		if !ok {
			return nil, Incompatible(path, services.KindLockit, "", "stem não é um kit de localização conhecido")
		}
		f, err := lockit.LoadFromBytes(layout, lang, data)
		if err != nil {
			return nil, err
		}
		var out []byte
		scope := builders.LockitApplyScope{
			Load: func(string) (*lockit.LockitFile, error) { return f, nil },
			Save: func(f *lockit.LockitFile) error {
				byLang, err := f.Bytes()
				if err != nil {
					return err
				}
				b, ok := byLang[lang]
				if !ok {
					return fmt.Errorf("lockit %s: idioma %s ausente no rebuild", layout.ID(), lang)
				}
				out = b
				return nil
			},
		}
		if err := builders.ApplyLockitDTOWithScope(version, c, scope); err != nil {
			return nil, err
		}
		if out == nil {
			return append([]byte(nil), data...), nil // nada mudou
		}
		return out, nil

	case services.KindHelp:
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		panel, err := helpPanelFromBytes(data, name, version, lang)
		if err != nil {
			return nil, err
		}
		var out []byte
		scope := builders.HelpApplyScope{
			Ids:   []string{name},
			Panel: func(string) *helpfile.HelpKeyedStringFile { return panel },
			Save: func(string) error {
				b, err := panel.ToBytes(lang)
				if err != nil {
					return err
				}
				out = b
				return nil
			},
		}
		if err := builders.ApplyHelpDTOWithScope(version, c, c.SortedKeys(), scope); err != nil {
			return nil, err
		}
		if out == nil {
			return append([]byte(nil), data...), nil
		}
		return out, nil

	default:
		return nil, Incompatible(path, "", d.Kind, "kind sem codec avulso")
	}
}

// applyWithTableScope roda o MESMO motor de apply de events/tabelas do app
// (builders.ApplyTextDTOWithScope — 4 fases, propagação restrita ao escopo)
// sobre os objetos decodificados do arquivo apontado. O Save captura os
// bytes reconstruídos no idioma do arquivo; sem mudança o ORIGINAL volta
// intacto (o import de um artefato que nada altera não regrava nada).
func applyWithTableScope(kind string, c dto.Collection, objs []*event.LocalizedFieldStringObject, version common.GameVersion, lang string, original []byte) ([]byte, error) {
	var out []byte
	scope := builders.TableApplyScope{
		Ids:        idsOf(c),
		StringsFor: func(string) []*event.LocalizedFieldStringObject { return objs },
		Save: func(string) error {
			b, err := event.EncodeLocalizedStrings(objs, lang, version)
			if err != nil {
				return err
			}
			out = b
			return nil
		},
	}
	if err := builders.ApplyTextDTOWithScope(kind, c, scope); err != nil {
		return nil, err
	}
	if out == nil {
		return append([]byte(nil), original...), nil // nada mudou
	}
	return out, nil
}

// parseFieldStringObjects decodifica o binário de events/tabela no idioma
// do arquivo — o objeto por posição carrega só esse idioma (export).
func parseFieldStringObjects(data []byte, version common.GameVersion, lang string) ([]*event.LocalizedFieldStringObject, error) {
	fs, err := event.FromFieldStringData(data, ffxencoding.GetCharsetForLanguage(lang), version)
	if err != nil {
		return nil, err
	}
	objs := make([]*event.LocalizedFieldStringObject, len(fs))
	for i, f := range fs {
		obj := event.NewLocalizedFieldStringObject()
		obj.SetLocalizedContent(lang, f)
		objs[i] = obj
	}
	return objs, nil
}

// parseFieldStringObjectsAllLangs decodifica com o idioma do arquivo e
// preenche placeholders vazios para os OUTROS idiomas: o motor de apply
// valida `GetLocalizedContent(lang) != nil` para cada idioma das rows —
// sem placeholder o artefato multi-idioma seria rejeitado. Os textos dos
// placeholders não são gravados (só o idioma do arquivo é codificado).
func parseFieldStringObjectsAllLangs(data []byte, version common.GameVersion, lang string) ([]*event.LocalizedFieldStringObject, error) {
	objs, err := parseFieldStringObjects(data, version, lang)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		for other := range common.SupportedLanguages {
			if other == lang {
				continue
			}
			obj.SetLocalizedContent(other, event.NewEmptyFieldString(ffxencoding.GetCharsetForLanguage(other), version))
		}
	}
	return objs, nil
}

// tableIDFor resolve o id de uma tabela events pelo caminho/nome (a mesma
// régua do matchVbfPath: cloudv/tutorial/menumain_btl têm ids próprios).
func tableIDFor(kind, path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	switch kind {
	case services.KindCloud:
		if strings.EqualFold(base, "cloudv") {
			return "cloudv"
		}
		return "cloud"
	case services.KindTutorial:
		if strings.EqualFold(base, "tutorial") {
			return "tutorial"
		}
		return "tuto0000"
	case services.KindMenuMain:
		if strings.Contains(strings.ToLower(strings.ReplaceAll(path, "\\", "/")), "/battle/btl/") {
			return "menumain_btl"
		}
		return "menumain"
	default:
		return base
	}
}

// objectSchemaFor resolve o schema chunkmap pelo NOME do arquivo dentro da
// versão: varre o registro por basename (o usuário aponta um arquivo
// avulso, não um caminho da árvore).
func objectSchemaFor(path string, version common.GameVersion) (string, chunkmap.SchemaBytesLoader, chunkmap.SchemaRebuilder, bool) {
	base := strings.ToLower(filepath.Base(path))
	prefix := common.VersionPathName(version) + "/"
	for key, load := range chunkmap.SchemaBytesLoaders {
		if !strings.HasPrefix(strings.ToLower(key), prefix) {
			continue
		}
		if strings.EqualFold(filepath.Base(key), base) {
			rebuild := chunkmap.SchemaRebuilders[key]
			return key, load, rebuild, true
		}
	}
	return "", nil, nil, false
}

// macroContainerFromBytes embrulha os bytes do macrodic.dcp apontado no
// container do idioma do arquivo (a localização é um arquivo só).
func macroContainerFromBytes(data []byte, version common.GameVersion, lang string) (map[string]*macrodic.MacroDictionaryBinaryFile, error) {
	c, err := macrodic.NewMacroDictionaryBinaryFileFromBytes(data, lang, version)
	if err != nil {
		return nil, err
	}
	return map[string]*macrodic.MacroDictionaryBinaryFile{lang: c}, nil
}

// lockitLayoutFor resolve o layout do lockit pelo stem do arquivo físico
// (<stem>_<lang>.bin): o id do layout é o stem sem o sufixo de idioma.
func lockitLayoutFor(path string, version common.GameVersion) (lockit.Layout, bool) {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	lang := lockitLanguageSuffix(strings.ToLower(base))
	stem := base
	if lang != "" {
		stem = base[:len(base)-len(lang)-1]
	}
	for _, l := range lockit.LayoutsForVersion(version) {
		if strings.EqualFold(l.ID(), stem) {
			return l, true
		}
	}
	return lockit.Layout{}, false
}

// helpPanelFromBytes decodifica o painel .sps2 apontado no objeto agregado
// que os builders consomem (uma localização carregada).
func helpPanelFromBytes(data []byte, name string, version common.GameVersion, lang string) (*helpfile.HelpKeyedStringFile, error) {
	bin, err := helpfile.ReadHelpBinary(data, name, lang, version)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	panel := helpfile.NewHelpKeyedStringFile(name)
	panel.Files[lang] = bin
	return panel, nil
}

// singleEntry devolve a única entrada da collection (o import avulso
// trabalha com a entrada casada por basename).
func singleEntry(c dto.Collection) (dto.FileEntry, string, error) {
	keys := c.SortedKeys()
	if len(keys) != 1 {
		return dto.FileEntry{}, "", fmt.Errorf("esperado 1 entrada no lote avulso, veio %d", len(keys))
	}
	return c[keys[0]], keys[0], nil
}

// idsOf devolve as chaves ordenadas (o escopo do motor de apply).
func idsOf(c dto.Collection) []string {
	return c.SortedKeys()
}

// detectFileLanguage adivinha o IDIOMA do arquivo avulso pelo caminho
// (new_<xx>pc/lockit _<lang>) — default: localização padrão (us).
func detectFileLanguage(path string) string {
	lower := strings.ToLower(strings.ReplaceAll(path, "\\", "/"))
	if idx := strings.Index(lower, "/new_"); idx >= 0 {
		rest := lower[idx+len("/new_"):]
		if len(rest) >= 2 {
			lang := rest[:2]
			if _, ok := common.SupportedLanguages[lang]; ok {
				return lang
			}
		}
	}
	base := strings.ToLower(filepath.Base(lower))
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	if lang := lockitLanguageSuffix(stem); lang != "" {
		return lang
	}
	return common.DefaultLocalization
}
