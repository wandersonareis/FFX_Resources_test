package chunkmap

import (
	"fmt"
	"sort"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/models"
)

// supportedLangs devolve os idiomas suportados em ordem — a mesma em que a
// extração e a importação iteram, para resultado determinístico.
func supportedLangs() []string {
	langs := make([]string, 0, len(common.SupportedLanguages))
	for locKey := range common.SupportedLanguages {
		langs = append(langs, locKey)
	}
	sort.Strings(langs)
	return langs
}

// TextResolver liga os campos de texto de uma struct à string table do
// arquivo. O campo é identificado pela sua chave — snake_case do nome do
// campo no struct C#, derivado por Fields/SegmentFields. Quem o implementa
// (MappedTextObject) é quem guarda o strtab e os segmentos resolvidos.
type TextResolver interface {
	// Texts devolve os textos não-vazios por idioma do campo `key`
	// (nil quando o campo não existe).
	Texts(key string) map[string]string
	// SetTexts grava `texts` no campo `key` e devolve o models.Segment
	// resultante (offset+chave do conteúdo atualizado na string table).
	SetTexts(key string, texts map[string]string) models.Segment
}

// ExportText devolve um FieldText por campo de texto de T, na ordem binária do
// arquivo (a mesma em que os textos aparecem no jogo, para o tradutor ler na
// mesma sequência). Campos sem texto são pulados.
//
// As chaves vêm de SegmentFields[T] — do próprio struct. Os textos vêm do
// resolver, que os lê dos slots (string table).
func ExportText[T any](r TextResolver) ([]objectsfile.FieldText, error) {
	fields, err := SegmentFields[T]()
	if err != nil {
		return nil, err
	}
	var out []objectsfile.FieldText
	for _, f := range fields {
		if texts := r.Texts(f.Key); len(texts) > 0 {
			out = append(out, objectsfile.FieldText{Key: f.Key, Texts: texts})
		}
	}
	return out, nil
}

// ImportText grava `fields` nos slots do resolver, na ordem em que
// aparecem no slice. Chave desconhecida é erro (pega typo no JSON); campo de
// `fields` vazio ou ausente deixa o slot como está.
//
// Os textos vivem nos slots do resolver (TextSlot.SetTexts codifica via
// converter): a struct recebe os offsets/chaves de volta em ToBytes, que é o
// ponto único de escrita e já pula slot sem conteúdo (bytes originais
// preservados).
func ImportText[T any](fields []objectsfile.FieldText, r TextResolver) error {
	textFields, err := SegmentFields[T]()
	if err != nil {
		return err
	}
	known := make(map[string]struct{}, len(textFields))
	for _, f := range textFields {
		known[f.Key] = struct{}{}
	}
	for _, f := range fields {
		if _, ok := known[f.Key]; !ok {
			return fmt.Errorf("chunkmap: campo de texto desconhecido %q", f.Key)
		}
		if len(f.Texts) == 0 {
			continue
		}
		r.SetTexts(f.Key, f.Texts)
	}
	return nil
}
