package objectsfile

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

// PointerStatus classifica um ref (offset na string table).
//
// O ponteiro é DIREÇÃO, não fato: o jogo guarda o texto em um bloco
// separado e o offset diz para ONDE olhar — ele só significa "há texto"
// quando cai no início de um bloco. Fora disso não existe texto ali, e o
// valor do ref jamais deve ser extraído (nem importado) como se fosse uma
// frase.
type PointerStatus uint8

const (
	// PointerCreated — ref criado em memória (apply num slot vazio), nunca
	// lido da tabela: não participa da validação da base.
	PointerCreated PointerStatus = iota
	// PointerOK — dentro da tabela, em fronteira de bloco, com texto.
	PointerOK
	// PointerEmpty — dentro da tabela, apontando para um terminador: o
	// "sem texto" legítimo. O jogo também não mostra nada ali.
	PointerEmpty
	// PointerMid — no MEIO de um texto anterior (o defeito de build_txt).
	PointerMid
	// PointerOOB — fora da string table.
	PointerOOB
)

func (s PointerStatus) String() string {
	switch s {
	case PointerCreated:
		return "created"
	case PointerEmpty:
		return "empty"
	case PointerMid:
		return "mid"
	case PointerOOB:
		return "oob"
	default:
		return "ok"
	}
}

// IsBroken diz se o ref aponta para o meio de outro texto ou para fora da
// tabela — o erro de binário. PointerEmpty não é erro: é o "sem texto"
// legítimo que o próprio arquivo já declara.
func (s PointerStatus) IsBroken() bool {
	return s == PointerMid || s == PointerOOB
}

// PointerStatusAt classifica `offset` contra a string table `table`.
//
// É a fronteira de bloco (o byte anterior ser terminador) que decide, não a
// ordem dos ponteiros: name_txt tem a row 0,6,10,13 estritamente crescente
// e ainda assim 10 cai dentro de "-----".
func PointerStatusAt(table []byte, offset int) PointerStatus {
	if offset < 0 || offset >= len(table) {
		return PointerOOB
	}
	if table[offset] == 0x00 {
		return PointerEmpty
	}
	if offset > 0 && table[offset-1] != 0x00 {
		return PointerMid
	}
	return PointerOK
}

// tableRef é um ref lido de uma string table: o status do offset contra a
// tabela da qual veio, os bytes armazenados e se o texto foi trocado depois.
// Implementado por *KeyedString (fluxo legado) e por chunkmap.TextContent —
// os dois domínios compartilham a mesma guarda.
type tableRef interface {
	PointerStatus() PointerStatus
	StoredBytes() []byte
	Edited() bool
	GetOffset() models.Offset
}

type KeyedString struct {
	Charset string
	Version common.GameVersion
	Segment models.Segment
	Bytes   []byte
	Text    string
	// Status é a classificação do Segment.Offset contra a tabela lida.
	// Refs com Status quebrado ficam SEM texto (Bytes vazio) e mantêm o
	// Segmento original no save.
	Status PointerStatus
	// edited marca texto trocado depois da leitura (SetString).
	edited bool
}

func NewKeyedString(charset string, segment models.Segment, data []byte, version common.GameVersion) datastore.IGlobalKeyedString {
	if segment.Offset == 0 && segment.Key == 0 {
		return nil
	}

	ks := &KeyedString{
		Charset: charset,
		Version: version,
		Segment: segment,
		Status:  PointerStatusAt(data, int(segment.Offset)),
	}

	// O ponteiro é direção, não fato: fora de uma fronteira de bloco (ou
	// fora da tabela) não há texto — não se lê o sufixo como se fosse frase.
	// PointerEmpty também fica sem bytes: já é o terminador.
	if ks.Status == PointerOK {
		ks.Bytes = converter.GetStringBytesAtLookupOffset(data, int(segment.Offset))
		ks.Text = converter.BytesToString(ks.Bytes, charset, version)
	}
	return ks
}

// PointerStatus devolve a classificação do ref lido (contrato tableRef).
func (ks *KeyedString) PointerStatus() PointerStatus { return ks.Status }

// StoredBytes devolve os bytes ARMAZENADOS do ref, sem terminador.
func (ks *KeyedString) StoredBytes() []byte { return ks.Bytes }

// Edited reporta se o texto do ref foi trocado depois da leitura.
func (ks *KeyedString) Edited() bool { return ks.edited }

func (ks *KeyedString) GetOffset() models.Offset {
	return ks.Segment.Offset
}

func (ks *KeyedString) SetOffset(offset models.Offset) {
	if offset != ks.Segment.Offset {
		ks.Segment.Offset = offset
	}
}

func (ks *KeyedString) GetKey() models.Key {
	return ks.Segment.Key
}

func (ks *KeyedString) SetKey(key models.Key) {
	if key != ks.Segment.Key {
		ks.Segment.Key = key
	}
}

func (ks *KeyedString) GetHeaderBytes(buf *bytes.Buffer) {
	binary.Write(buf, binary.LittleEndian, uint16(ks.Segment.Offset))
	binary.Write(buf, binary.LittleEndian, uint16(ks.Segment.Key))
}

func (ks *KeyedString) SetHeaderBytes(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	if len(ks.Bytes) == 0 {
		ks.Bytes = make([]byte, 0)
	}
	if ks.Segment.Offset == 0 && ks.Segment.Key == 0 {
		return
	}
	/* binary.Write(buf, binary.LittleEndian, uint16(ks.Segment.Offset))
	binary.Write(buf, binary.LittleEndian, uint16(ks.Segment.Key)) */
	if err := models.WriteSegment(buf, ks.Segment); err != nil {
		return
	}
}

func (ks *KeyedString) GetCharset() string {
	if ks.Charset == "" {
		return common.DefaultLocalization
	}
	return ks.Charset
}

func (ks *KeyedString) SetCharset(charset string) {
	if charset != "" && charset != ks.Charset {
		ks.Charset = charset
	}
}

func (ks *KeyedString) String() string {
	return ks.GetString()
}

func (ks *KeyedString) GetString() string {
	return converter.BytesToString(ks.Bytes, ks.Charset, ks.Version)
}

func (ks *KeyedString) IsEmpty() bool {
	return ks.GetString() == ""
}

func (ks *KeyedString) SetString(str, newCharset string) {
	if newCharset != "" && newCharset != ks.Charset {
		ks.Charset = newCharset
	}
	encoded, err := converter.StringToBytes(str, ks.Charset, ks.Version)
	if err != nil {
		common.LogError("SetString: %v", err)
		return
	}
	ks.Bytes = encoded
	ks.edited = true
}

// RebuildKeyedStrings compila os textos na string table e devolve os bytes
// dela, recalculando os offsets dos refs com texto.
//
// `base` é a string table ORIGINAL do arquivo. Quando ela reproduz os refs
// do arquivo, o resultado é montado como BASE + APPEND: os offsets
// originais continuam apontando para o mesmo lugar e os refs sem texto
// (MID/OOB/EMPTY) mantêm o segmento cru byte a byte. Quando a base não
// serve, o rebuild dedup de sempre roda intocado — nada muda onde não há
// problema.
//
// Serve os dois domínios (KeyedString legado e chunkmap.TextContent): a
// distinção está no contrato tableRef, não no tipo.
func RebuildKeyedStrings(list []datastore.IGlobalKeyedString, charset string, version common.GameVersion, base []byte) []byte {
	if base = UsableStringTableBase(list, base); len(base) > 0 {
		return rebuildOnStringTableBase(list, base, charset, version)
	}
	return rebuildStringTableFromScratch(list, charset, version)
}

// UsableStringTableBase decide se `candidate` pode servir de base do
// rebuild. A base só é aceita quando reproduz todos os refs LIDOS (mesmo
// offset, mesmos bytes / mesma ausência de texto) e ainda prova pelo menos
// UM ref quebrado — é a prova de que ela é a tabela da qual os offsets
// vieram. Refs editados ou criados em memória não provam nem refutam.
// Arquivo limpo devolve nil.
func UsableStringTableBase(list []datastore.IGlobalKeyedString, candidate []byte) []byte {
	if len(candidate) == 0 {
		return nil
	}

	proved := false
	for _, ks := range list {
		if ks == nil {
			continue
		}
		ref, ok := ks.(tableRef)
		if !ok {
			// Ref de outro domínio: não dá para validar contra a tabela.
			return nil
		}
		if ref.Edited() {
			continue
		}
		status := ref.PointerStatus()
		if status == PointerCreated {
			continue
		}

		off := int(ref.GetOffset())
		switch status {
		case PointerOK:
			if PointerStatusAt(candidate, off) != PointerOK {
				return nil
			}
			if raw := ref.StoredBytes(); len(raw) > 0 &&
				!bytes.Equal(converter.GetStringBytesAtLookupOffset(candidate, off), raw) {
				return nil
			}
		case PointerEmpty:
			if PointerStatusAt(candidate, off) != PointerEmpty {
				return nil
			}
		default: // PointerMid, PointerOOB
			if PointerStatusAt(candidate, off) != status {
				return nil
			}
			proved = true
		}
	}
	if !proved {
		return nil
	}
	return candidate
}

// appendedRef guarda um ref cujo texto foi escrito NOVAMENTE (não existia na
// base, ou o texto mudou): o offset dele é o início do bloco dentro da cauda,
// contado a partir do fim da base.
type appendedRef struct {
	ks   datastore.IGlobalKeyedString
	tail int
}

// rebuildOnStringTableBase monta a tabela como base + cauda. Nada que já está
// no arquivo é reescrito: o offset original só muda quando o texto do ref
// mudou (e mesmo assim reaproveita bytes já presentes quando pode).
func rebuildOnStringTableBase(list []datastore.IGlobalKeyedString, base []byte, charset string, version common.GameVersion) []byte {
	baseLen := len(base)
	baseIndex := indexStringTable(base)

	var (
		tail         bytes.Buffer
		tailIndex    = make(map[string]models.Offset)
		appended     []appendedRef
		preservedOOB []int
	)

	for _, ks := range list {
		if ks == nil {
			continue
		}
		off := int(ks.GetOffset())

		// Ref sem texto fora de uma fronteira de bloco (MID/OOB) ou apontando
		// para um terminador (EMPTY): o ponteiro original É o valor do
		// arquivo, e é ele que o jogo lê como "não tem texto". Não recalcula.
		if ks.GetString() == "" {
			if status := PointerStatusAt(base, off); status != PointerOK {
				if status == PointerOOB {
					preservedOOB = append(preservedOOB, off)
				}
				continue
			}
		}

		raw := storedBytesOf(ks, charset, version)

		// Texto já gravado exatamente onde o ref aponta: offset intocado.
		if len(raw) > 0 && bytes.Equal(converter.GetStringBytesAtLookupOffset(base, off), raw) {
			continue
		}

		// Reaproveita o que já existe — primeiro na base, depois na cauda.
		key := string(raw)
		if shared, exists := baseIndex[key]; exists {
			ks.SetOffset(shared)
			continue
		}
		if shared, exists := tailIndex[key]; exists {
			appended = append(appended, appendedRef{ks: ks, tail: int(shared)})
			continue
		}

		tailPos := tail.Len()
		tailIndex[key] = models.Offset(tailPos)
		tail.Write(raw)
		tail.WriteByte(0x00)
		appended = append(appended, appendedRef{ks: ks, tail: tailPos})
	}

	// O arquivo só cresce quando há texto novo, e um ref OOB preservado
	// (>= baseLen) é justamente o que o crescimento ameaça: se ele cair
	// numa fronteira de bloco novo vira ponteiro VÁLIDO e o jogo passa a
	// exibir texto onde o arquivo original não tinha. O padding de 0x00
	// resolve sem reescrever byte nenhum — a posição vira terminador ou cai
	// no meio de um bloco; nos dois casos, sem texto.
	padding := 0
	if tail.Len() > 0 && len(preservedOOB) > 0 {
		padding = safeTailPadding(base, preservedOOB, tail.Bytes())
	}

	var out bytes.Buffer
	out.Grow(baseLen + padding + tail.Len())
	out.Write(base)
	out.Write(make([]byte, padding))
	out.Write(tail.Bytes())

	start := baseLen + padding
	for _, a := range appended {
		a.ks.SetOffset(models.Offset(start + a.tail))
	}
	return out.Bytes()
}

// safeTailPadding devolve quantos 0x00 inserir entre a base e a cauda para
// que nenhum offset OOB preservado caia na fronteira de um bloco novo.
//
// Os valores "piores" são finitos — um por par (OOB, fronteira) — então o
// primeiro m não-pior é menor que a contagem deles: o padding nunca passa
// de alguns bytes.
func safeTailPadding(base []byte, oobOffsets []int, tail []byte) int {
	blocks := blockStarts(tail)
	bad := make(map[int]bool, len(oobOffsets)*len(blocks))
	for _, o := range oobOffsets {
		for _, b := range blocks {
			if d := o - len(base) - b; d >= 0 {
				bad[d] = true
			}
		}
	}
	for m := 0; ; m++ {
		if !bad[m] {
			return m
		}
	}
}

// blockStarts devolve as posições em que um bloco começa: o início do trecho
// e todo byte imediatamente depois de um terminador.
func blockStarts(tail []byte) []int {
	starts := []int{0}
	for i := 1; i < len(tail); i++ {
		if tail[i-1] == 0x00 {
			starts = append(starts, i)
		}
	}
	return starts
}

// rebuildStringTableFromScratch é o rebuild dedup (arquivos sem nenhum ref
// quebrado): compila os textos do zero, compartilhando offsets entre
// repetições e entre campos vazios.
func rebuildStringTableFromScratch(list []datastore.IGlobalKeyedString, charset string, version common.GameVersion) []byte {
	var buf bytes.Buffer
	// Dedup de bytes: texto já gravado no arquivo compartilha o offset
	// (mesma técnica do RebuildFieldStrings dos events) — kernel com
	// repetição massiva intra-arquivo (ex.: "Attack" centenas de vezes)
	// compila sem duplicar bytes. O arquivo continua existindo inteiro;
	// a deduplicação de trabalho é para o tradutor, a de offsets é de
	// formato. Texto VAZIO também deduplica: todos os campos vazios
	// compartilham um único 0x00 (o original não desperdiça byte por
	// campo vazio — aponta todos para o mesmo terminador).
	offsetMap := make(map[string]models.Offset)

	for _, ks := range list {
		if ks == nil {
			continue
		}
		s := ks.GetString()

		if offset, shared := offsetMap[s]; shared {
			ks.SetOffset(offset)
			continue
		}
		offsetMap[s] = models.Offset(buf.Len())
		ks.SetOffset(models.Offset(buf.Len()))
		converter.FillByteList(s, &buf, charset, version)
	}

	return buf.Bytes()
}

// indexStringTable mapeia os bytes de cada bloco da tabela (sem terminador)
// para o offset do bloco. É por bytes, não por texto decodificado: apontar
// para o bloco original é a única forma de o jogo ler exatamente o que o
// arquivo já declara.
func indexStringTable(base []byte) map[string]models.Offset {
	index := make(map[string]models.Offset)
	for off := 0; off < len(base); {
		end := bytes.IndexByte(base[off:], 0x00)
		if end < 0 {
			registerOnce(index, base[off:], off)
			break
		}
		registerOnce(index, base[off:off+end], off)
		off += end + 1
	}
	return index
}

func registerOnce(index map[string]models.Offset, raw []byte, off int) {
	key := string(raw)
	if _, exists := index[key]; !exists {
		index[key] = models.Offset(off)
	}
}

// storedBytesOf devolve os bytes ARMAZENADOS do ref (sem terminador).
func storedBytesOf(ks datastore.IGlobalKeyedString, charset string, version common.GameVersion) []byte {
	if ref, ok := ks.(tableRef); ok {
		return ref.StoredBytes()
	}
	raw, err := converter.StringToBytes(ks.GetString(), charset, version)
	if err != nil {
		common.LogError("storedBytesOf: %v", err)
		return nil
	}
	return raw
}

// PointerReport quantifica os refs de um arquivo por status. É o
// diagnóstico que substitui o drop silencioso de antes: os refs quebrados
// não viram linha de texto, e aqui se mede quantos existem.
type PointerReport struct {
	OK      int
	Created int
	Empty   int
	Mid     int
	OOB     int
}

// Broken é a soma de MID + OOB: refs que apontam para o meio de outro
// texto ou para fora da tabela.
func (r PointerReport) Broken() int {
	return r.Mid + r.OOB
}

// ReportKeyedStrings classifica cada ref de uma lista de conteúdo lido.
func ReportKeyedStrings(list []datastore.IGlobalKeyedString) PointerReport {
	var r PointerReport
	for _, ks := range list {
		ref, ok := ks.(tableRef)
		if !ok {
			continue
		}
		switch ref.PointerStatus() {
		case PointerCreated:
			r.Created++
		case PointerEmpty:
			r.Empty++
		case PointerMid:
			r.Mid++
		case PointerOOB:
			r.OOB++
		default:
			r.OK++
		}
	}
	return r
}
