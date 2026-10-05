package objectsfile

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
	"sort"
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
	// Refs com Status quebrado ficam SEM texto (Bytes vazio) e no save só
	// acompanham o deslocamento do splice — nunca são reescritos.
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
// dela, recalculando os offsets dos refs.
//
// `base` é a string table ORIGINAL do arquivo, e é ela a fonte da ordem: o
// rebuild faz SPLICE do texto editado no próprio bloco e desloca todo
// offset situado depois pela soma das crescidas anteriores (deslocamento
// global). Refs MID/OOB/EMPTY não têm bytes — recebem deslocamento sem
// causar deslocamento. Sem edição o delta é zero em todo ponto e a saída é
// o base byte a byte.
//
// A base só é usada quando reproduz os refs lidos (UsableStringTableBase);
// sem candidato utilizável cai no rebuild dedup de sempre.
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
// rebuild. A base só é aceita quando reproduz todos os refs LIDOS: mesmo
// offset, mesmos bytes e mesma classificação (inclusive os MID/OOB, que
// têm que continuar quebrados contra ela). É a prova de que `candidate` é
// a tabela da qual os offsets vieram. Refs editados ou criados em memória
// não participam da validação.
//
// NÃO depende de haver ref quebrado: o splice usa a base no arquivo limpo
// também — sem texto editado o delta é zero em todo ponto e a saída é o
// próprio base byte a byte. Sem candidato, devolve nil.
func UsableStringTableBase(list []datastore.IGlobalKeyedString, candidate []byte) []byte {
	if len(candidate) == 0 {
		return nil
	}

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
		}
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

// splicePoint descreve a substituição de UM bloco da string table original:
// o trecho [off, off+oldLen) sai e `newRaw` + terminador entra no lugar.
type splicePoint struct {
	off    int
	oldLen int
	newRaw []byte
}

// originalBlockLen devolve o tamanho do bloco (texto + terminador) que
// começa em `off` na tabela original.
//
// É de onde vem o tamanho que o splice substitui: SetString apaga os bytes
// originais do ref, mas a tabela original ainda traz o bloco inteiro.
func originalBlockLen(base []byte, off int) int {
	if off < 0 || off >= len(base) {
		return 0
	}
	end := bytes.IndexByte(base[off:], 0x00)
	if end < 0 {
		return len(base) - off
	}
	return end + 1 // inclui o terminador
}

// rebuildOnStringTableBase monta a string table por SPLICE sobre o base
// original — um caminho só, para arquivo limpo e para arquivo com ref
// quebrado.
//
// Regras:
//
//   - ref PointerOK EDITADO    -> o próprio bloco é substituído no lugar;
//     o delta (len(novo) - len(velho)) passa a valer a partir dali.
//   - ref PointerOK intocado   -> só desloca.
//   - ref MID / OOB / EMPTY    -> só deslocam. Não têm bytes nem texto:
//     RECEBEM deslocamento sem CAUSAR deslocamento.
//   - ref criado em memória (ou editado a partir de um offset que não é o
//     seu bloco) -> vai para a cauda: não existe posição no original.
//
// Todo offset situado depois de um splice desloca pela soma das crescidas
// anteriores a ele — deslocamento global, não só do ref editado. Um MID
// dentro do bloco spliceado cai junto com ele, e um OOB (>= len(base)) soma
// o delta total, de modo que len_novo - off_novo = len_base - off_base: ele
// continua fora da tabela sem precisar de guarda extra.
//
// Sem texto editado o delta é 0 em todo ponto e a saída É o base, byte a
// byte.
func rebuildOnStringTableBase(list []datastore.IGlobalKeyedString, base []byte, charset string, version common.GameVersion) []byte {
	// Fotografia dos offsets ORIGINAIS: os SetOffset do fim não podem
	// influenciar a própria contagem de deslocamento.
	type placement struct {
		ks      datastore.IGlobalKeyedString
		origOff int
	}

	var (
		shift       []placement
		cauda       []datastore.IGlobalKeyedString
		grupos      = map[int][]datastore.IGlobalKeyedString{}
		naoEditados = map[int]bool{}
	)

	for _, ks := range list {
		if ks == nil {
			continue
		}
		ref, ok := ks.(tableRef)
		if !ok {
			continue // base não teria sido aceita, mas não se arrisca
		}
		off := int(ks.GetOffset())
		switch st := ref.PointerStatus(); {
		case st == PointerCreated:
			cauda = append(cauda, ks)
		case !ref.Edited():
			// Todo offset ocupado por um ref intocado é intocável: o
			// splice mudaria o texto dele sem querer.
			naoEditados[off] = true
			shift = append(shift, placement{ks: ks, origOff: off})
		case st == PointerOK:
			// Dedup: vários refs podem apontar o MESMO bloco — ele é
			// substituído uma vez só.
			grupos[off] = append(grupos[off], ks)
		default:
			// Editado partindo de um offset que não é o seu bloco:
			// nunca se splica no meio do texto de outro.
			cauda = append(cauda, ks)
		}
	}

	ordem := make([]int, 0, len(grupos))
	for off := range grupos {
		ordem = append(ordem, off)
	}
	sort.Ints(ordem)

	var (
		splices   []splicePoint
		splicados []placement
	)
	for _, off := range ordem {
		refs := grupos[off]
		oldLen := originalBlockLen(base, off)
		if naoEditados[off] || oldLen == 0 {
			// Há um ref NÃO editado apontando o mesmo bloco (o splice
			// mudaria o texto dele sem querer), ou o offset não delimita
			// bloco: os editados vão para a cauda.
			cauda = append(cauda, refs...)
			continue
		}
		raw := storedBytesOf(refs[0], charset, version)
		splices = append(splices, splicePoint{off: off, oldLen: oldLen, newRaw: raw})
		for _, ks := range refs {
			splicados = append(splicados, placement{ks: ks, origOff: off})
		}
	}

	// Delta acumulado ANTES de cada splice: prefixo[i] soma os deltas de
	// splices[0..i-1], então deltaAntes(p) soma os de off < p.
	spliceOffs := make([]int, len(splices))
	prefix := make([]int, len(splices)+1)
	for i, s := range splices {
		spliceOffs[i] = s.off
		prefix[i+1] = prefix[i] + (len(s.newRaw) + 1 - s.oldLen)
	}
	deltaAntes := func(p int) int {
		return prefix[sort.SearchInts(spliceOffs, p)]
	}

	// O corpo: base com cada bloco editado trocado no lugar.
	var out bytes.Buffer
	out.Grow(len(base))
	cursor := 0
	for _, s := range splices {
		if s.off < cursor {
			continue // sobreposição não ocorre: blocos são disjuntos
		}
		out.Write(base[cursor:s.off])
		out.Write(s.newRaw)
		out.WriteByte(0x00)
		cursor = s.off + s.oldLen
	}
	out.Write(base[cursor:])

	// Reposiciona tudo: o que vem depois de um splice desloca pelo delta
	// acumulado das crescidas anteriores.
	for _, p := range shift {
		p.ks.SetOffset(models.Offset(p.origOff + deltaAntes(p.origOff)))
	}
	for _, p := range splicados {
		p.ks.SetOffset(models.Offset(p.origOff + deltaAntes(p.origOff)))
	}

	// Cauda só para quem não tem posição no original.
	var (
		tail         bytes.Buffer
		tailIndex    = make(map[string]models.Offset)
		appended     []appendedRef
		preservedOOB []int
	)
	for _, ks := range cauda {
		raw := storedBytesOf(ks, charset, version)
		key := string(raw)
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
	for _, p := range shift {
		if r, ok := p.ks.(tableRef); ok && r.PointerStatus() == PointerOOB {
			preservedOOB = append(preservedOOB, p.origOff+deltaAntes(p.origOff))
		}
	}

	// A cauda é a única região que cresce SEM que o offset OOB acompanhe
	// (ele já está deslocado até o fim do corpo): se cair na fronteira de
	// um bloco novo vira ponteiro VÁLIDO. O padding de 0x00 resolve — a
	// posição vira terminador ou cai no meio de um bloco; sem texto nos
	// dois casos.
	padding := 0
	if tail.Len() > 0 && len(preservedOOB) > 0 {
		padding = safeTailPadding(out.Bytes(), preservedOOB, tail.Bytes())
	}

	start := out.Len() + padding
	result := make([]byte, 0, out.Len()+padding+tail.Len())
	result = append(result, out.Bytes()...)
	result = append(result, make([]byte, padding)...)
	result = append(result, tail.Bytes()...)

	for _, a := range appended {
		a.ks.SetOffset(models.Offset(start + a.tail))
	}
	return result
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
