// Package weaponnames é a política de domínio de battle/kernel/w_name.bin no
// fluxo chunkmap (o equivalente a weapon_names_object.go no fluxo antigo).
//
// w_name.bin é EXCLUSIVO do FFX: os personagens são fixos (Tidus, Yuna,
// Auron, Kimahri, Wakka, Lulu, Rikku) e não existe possibilidade de adicionar
// ou remover personagens — nem compatibilidade com FFX-2, que tem outros
// personagens em outro arquivo. A tabela model_id -> dono em
// schema/ffx/wname.go é, portanto, ESTÁTICA: qualquer par de bytes fora dela
// é erro e aborta o processamento (nunca fallback posicional).
//
// O motor chunkmap não precisa de nenhuma exceção para este arquivo: a struct
// ffx.WeaponName mapeia os 72 bytes do chunk sem saltos. O que é exclusivo de
// w_name é só a apresentação/identidade das rows — este pacote:
//
//   - exporta cada slot rotulado pelo dono resolvido nos model_ids do próprio
//     chunk ("Tidus.Standard"), com Character/Kind como referência de exibição
//     (somente os NOMES de armas/escudos são editáveis; o model_id nunca vira
//     row nem campo do JSON);
//   - importa por NOME, localizando a posição atual do personagem pelos
//     model_ids do registro — a posição no array nunca é identidade;
//   - trata a leitura em hiragina (names_hira): como os arrays Standard e
//     Simplified não são adjacentes, o fallback por SameAs do TextPair não
//     alcança — row Simplified ausente no JSON significa "leitura igual ao
//     nome" e o import replica o Standard.
package weaponnames

import (
	"fmt"
	"path/filepath"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/fileFormats/objectsfile/chunkmap"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx"
	"ffxresources/backend/formatters/hash"
	"ffxresources/backend/models"
)

// Pattern é o pattern path do binário de nomes de equipamentos (FFX only).
const Pattern = "battle/kernel/w_name.bin"

// characterOrder é a ordem fixa dos personagens por slot no arquivo de
// referência (T/Y/A/K/W/L/R). Usada APENAS como assert de integridade do
// binário pristine (AssertColumnOrder, em teste) — nunca como identidade no
// export/import, que sempre resolve pelos model_ids do chunk.
var characterOrder = [...]string{
	"Tidus", "Yuna", "Auron", "Kimahri", "Wakka", "Lulu", "Rikku",
}

// Slots devolve os índices das colunas textuais Standard (names) e
// Simplified (names_hira) de cada personagem (0..6), derivados dos
// metadados do schema — nada hardcoded além da contagem esperada.
func Slots() (std [7]int, simp [7]int, err error) {
	fields, err := chunkmap.SegmentFields[ffx.WeaponName]()
	if err != nil {
		return std, simp, err
	}
	if len(fields) != 2*len(characterOrder) {
		return std, simp, fmt.Errorf("w_name: %d campos textuais, esperado %d", len(fields), 2*len(characterOrder))
	}
	for i, field := range fields {
		elem := -field.Path[len(field.Path)-1] - 1
		if elem < 0 || elem >= len(characterOrder) {
			return std, simp, fmt.Errorf("w_name: elemento fora da faixa em %s", field.Name)
		}
		switch field.Part {
		case models.TextPartStandard:
			std[elem] = i
		case models.TextPartSimplified:
			simp[elem] = i
		default:
			return std, simp, fmt.Errorf("w_name: parte textual inesperada em %s", field.Name)
		}
	}
	return std, simp, nil
}

// Validate confere os model_ids de TODOS os chunks contra a tabela estática
// de schema/ffx/wname.go. ID sem dono ou personagem repetido no mesmo chunk
// tornam a identidade das rows ambígua — erro e aborte, nunca fallback.
func Validate(file *chunkmap.File[ffx.WeaponName]) error {
	if file == nil {
		return fmt.Errorf("w_name: arquivo ausente")
	}
	for chunk := 0; chunk < file.Count(); chunk++ {
		seen := make(map[string]struct{}, len(characterOrder))
		for slot, id := range file.Get(chunk).Value().ModelIDs {
			key := ffx.WeaponModelKeyOf(id)
			owner, ok := ffx.WeaponModelOwnerOf(key)
			if !ok {
				return fmt.Errorf("w_name: chunk %d, slot %d: model id %s sem dono no mapa — estenda schema/ffx/wname.go",
					chunk, slot, key)
			}
			if _, duplicate := seen[owner.Character]; duplicate {
				return fmt.Errorf("w_name: chunk %d: model id %s repete o personagem %s", chunk, key, owner.Character)
			}
			seen[owner.Character] = struct{}{}
		}
	}
	return nil
}

// AssertColumnOrder prova que o arquivo pristine segue a ordem fixa de
// characterOrder (slot i = personagem i). É uma garantia sobre o binário de
// referência — não uma restrição do fluxo, que aceita chunks reordenados
// desde que cada model_id resolva dono único (ver Validate).
func AssertColumnOrder(file *chunkmap.File[ffx.WeaponName]) error {
	for chunk := 0; chunk < file.Count(); chunk++ {
		for slot, id := range file.Get(chunk).Value().ModelIDs {
			owner, ok := ffx.WeaponModelOwnerOf(ffx.WeaponModelKeyOf(id))
			if !ok {
				return fmt.Errorf("w_name: chunk %d, slot %d: model id sem dono", chunk, slot)
			}
			if owner.Character != characterOrder[slot] {
				return fmt.Errorf("w_name: chunk %d, slot %d: dono %s, esperado %s (ordem de coluna quebrada)",
					chunk, slot, owner.Character, characterOrder[slot])
			}
		}
	}
	return nil
}

// Rows monta as rows de export do fluxo novo. O dono de cada slot vem do
// mapa de model IDs (Character + Kind como REFERÊNCIA de exibição); ID fora
// da tabela é erro — estenda schema/ffx/wname.go, nunca presuma posição. A
// row Simplified só existe quando a leitura difere do nome em algum idioma
// (mesma regra do TextPair — ver doc do pacote).
func Rows(defaultFile *chunkmap.File[ffx.WeaponName], files map[string]*chunkmap.File[ffx.WeaponName]) ([]dto.TextRow, error) {
	if err := Validate(defaultFile); err != nil {
		return nil, err
	}
	std, simp, err := Slots()
	if err != nil {
		return nil, err
	}

	type fieldAddress struct{ chunk, field int }
	texts := make(map[fieldAddress]map[string]string)
	for _, language := range exportLanguages(files) {
		file := files[language]
		if file == nil {
			file = defaultFile
		}
		exported, err := file.ExportText(language)
		if err != nil {
			return nil, fmt.Errorf("w_name/%s: exportar textos: %w", language, err)
		}
		for _, chunk := range exported {
			for _, field := range chunk.Fields {
				address := fieldAddress{chunk: chunk.Index, field: field.Index}
				merged := texts[address]
				if merged == nil {
					merged = make(map[string]string)
					texts[address] = merged
				}
				merged[language] = field.Texts[language]
			}
		}
	}

	rows := make([]dto.TextRow, 0, defaultFile.Count()*len(characterOrder))
	for chunk := 0; chunk < defaultFile.Count(); chunk++ {
		for elem, id := range defaultFile.Get(chunk).Value().ModelIDs {
			owner, ok := ffx.WeaponModelOwnerOf(ffx.WeaponModelKeyOf(id))
			if !ok {
				return nil, fmt.Errorf("w_name: chunk %d: model id sem dono no mapa", chunk)
			}
			standard := texts[fieldAddress{chunk: chunk, field: std[elem]}]
			if len(standard) > 0 {
				rows = append(rows, row(chunk, owner, models.TextPartStandard, standard))
			}
			hira := texts[fieldAddress{chunk: chunk, field: simp[elem]}]
			if len(hira) > 0 && differs(hira, standard) {
				rows = append(rows, row(chunk, owner, models.TextPartSimplified, hira))
			}
		}
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("w_name: nenhum texto para exportar")
	}
	return rows, nil
}

// Apply grava as rows de volta no arquivo. O import lida só com os NOMES
// ("Tidus.Standard"/"Tidus.Simplified"): a posição do personagem é resolvida
// pelos model_ids do próprio registro (nunca pela ordem) e row Simplified
// ausente replica o texto do Standard nos idiomas em que a leitura atual tem
// texto (leitura vazia não é preenchida — ref quebrado mantém o segmento
// original).
func Apply(file *chunkmap.File[ffx.WeaponName], rows []dto.TextRow) error {
	if err := Validate(file); err != nil {
		return err
	}
	std, simp, err := Slots()
	if err != nil {
		return err
	}
	fields, err := chunkmap.SegmentFields[ffx.WeaponName]()
	if err != nil {
		return err
	}

	type ownerKey struct {
		chunk     int
		character string
	}
	type ownerRows struct {
		standard, simplified map[string]string
	}
	byOwner := make(map[ownerKey]*ownerRows)
	for _, r := range rows {
		if r.Index < 0 || r.Index >= file.Count() {
			return fmt.Errorf("w_name: chunk %d fora da faixa (%d chunks)", r.Index, file.Count())
		}
		dot := strings.LastIndex(r.Name, ".")
		if dot <= 0 || dot == len(r.Name)-1 {
			return fmt.Errorf("w_name: row %d sem variante no nome %q", r.Index, r.Name)
		}
		character, variant := r.Name[:dot], r.Name[dot+1:]
		if variant != string(models.TextPartStandard) && variant != string(models.TextPartSimplified) {
			return fmt.Errorf("w_name: row %d com variante desconhecida %q", r.Index, variant)
		}
		target := byOwner[ownerKey{r.Index, character}]
		if target == nil {
			target = &ownerRows{}
			byOwner[ownerKey{r.Index, character}] = target
		}
		switch variant {
		case string(models.TextPartStandard):
			if target.standard != nil {
				return fmt.Errorf("w_name: chunk %d: row %s repetida", r.Index, r.Name)
			}
			target.standard = r.Text
		case string(models.TextPartSimplified):
			if target.simplified != nil {
				return fmt.Errorf("w_name: chunk %d: row %s repetida", r.Index, r.Name)
			}
			target.simplified = r.Text
		}
	}

	for owner, target := range byOwner {
		pos := -1
		for i, id := range file.Get(owner.chunk).Value().ModelIDs {
			resolved, ok := ffx.WeaponModelOwnerOf(ffx.WeaponModelKeyOf(id))
			if ok && resolved.Character == owner.character {
				if pos >= 0 {
					return fmt.Errorf("w_name: chunk %d: personagem %s aparece em dois model ids", owner.chunk, owner.character)
				}
				pos = i
			}
		}
		if pos < 0 {
			return fmt.Errorf("w_name: chunk %d: personagem %s ausente dos model_ids", owner.chunk, owner.character)
		}
		if target.standard != nil {
			if err := applySlot(file, owner.chunk, fields[std[pos]].Name, std[pos], target.standard); err != nil {
				return err
			}
		}
		hira := target.simplified
		if hira == nil && target.standard != nil {
			hira = make(map[string]string, len(target.standard))
			for language, text := range target.standard {
				if current := currentText(file, owner.chunk, simp[pos], language); current != "" {
					hira[language] = text
				}
			}
			if len(hira) == 0 {
				hira = nil
			}
		}
		if hira != nil {
			if err := applySlot(file, owner.chunk, fields[simp[pos]].Name, simp[pos], hira); err != nil {
				return err
			}
		}
	}
	return nil
}

// Collection monta a dto.Collection de w_name.bin para os formatters. As
// chaves seguem os padroes dos demais objetos: entry = basename sem extensao
// e metadata.key = FileLayoutKey - aceito pelo pipeline de import do app.
// aceito pelo pipeline do app sem traducao.
func Collection(version common.GameVersion, rows []dto.TextRow) (dto.Collection, error) {
	key := strings.TrimSuffix(filepath.Base(Pattern), filepath.Ext(Pattern))
	entry := dto.FileEntry{
		Metadata: dto.NewObjectMetadata(
			version,
			filepath.ToSlash(filepath.Dir(Pattern)),
			filepath.Base(Pattern),
			objectsfile.FileLayoutKey(version, Pattern),
		).WithRowCount(len(rows)),
		Rows: rows,
	}
	collection := dto.Collection{key: entry}
	if err := hash.ValidateNoCollision(collection); err != nil {
		return nil, err
	}
	return collection, nil
}

// row monta uma row de export: Index+Name são a identidade; Character/Kind
// são referência de exibição (o model_id não entra no artefato).
func row(chunk int, owner ffx.WeaponModelOwner, part models.TextPart, texts map[string]string) dto.TextRow {
	return dto.TextRow{
		Index:     chunk,
		Character: owner.Character,
		Kind:      owner.Kind,
		Name:      owner.Character + "." + string(part),
		Hash:      hash.Texts(texts),
		Text:      texts,
	}
}

// differs reporta se a leitura difere do nome em algum idioma — igual em tudo
// não vira row própria (o import replica o Standard).
func differs(hira, standard map[string]string) bool {
	for language, text := range hira {
		if standard[language] != text {
			return true
		}
	}
	return false
}

// exportLanguages devolve os idiomas a exportar: os presentes no mapa de
// arquivos carregados por locale, mais o default SEMPRE (sem arquivo próprio
// no mapa, o próprio defaultFile é usado).
func exportLanguages(files map[string]*chunkmap.File[ffx.WeaponName]) []string {
	out := make([]string, 0, len(files)+1)
	seen := make(map[string]struct{}, len(files)+1)
	for language := range files {
		if _, duplicate := seen[language]; !duplicate {
			seen[language] = struct{}{}
			out = append(out, language)
		}
	}
	if _, ok := seen[common.DefaultLocalization]; !ok {
		out = append(out, common.DefaultLocalization)
	}
	return out
}

// currentText devolve o texto ATUAL do slot no idioma ("" quando o slot não
// tem conteúdo — ExportText omite textos vazios).
func currentText(file *chunkmap.File[ffx.WeaponName], chunk, slot int, language string) string {
	exported, err := file.Get(chunk).ExportText(language)
	if err != nil {
		return ""
	}
	for _, field := range exported {
		if field.Index == slot {
			return field.Texts[language]
		}
	}
	return ""
}

func applySlot(file *chunkmap.File[ffx.WeaponName], chunk int, name string, slot int, texts map[string]string) error {
	return file.Get(chunk).ImportText([]chunkmap.TextField{{Index: slot, Name: name, Texts: texts}})
}
