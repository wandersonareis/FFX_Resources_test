package ddsphyre

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"sync"

	"ffxresources/backend/common"
)

// errMissing marca "não existe nesta fonte" (o índice pula o arquivo).
var errMissing = errors.New("arquivo não encontrado")

// DUPLICATAS (.dds.phyre idênticos)
//
// A árvore do jogo tem ~2500 texturas mas só ~700 imagens distintas: a
// otimização do DVD replicou o mesmo conteúdo por caminho (como nos textos
// repetidos). Encontrar essas cópias exige hashear o PAYLOAD, não o arquivo:
//
//   - hash do arquivo inteiro encontra quase nada (2523 distintos em 2527) —
//     o namespace embute o nome/caminho do próprio arquivo, então duas
//     cópias idênticas diferem em 1-2 bytes, os dígitos do nome;
//   - hash do payload (raw[dataOffset:], os pixels puros, sem header DDS)
//     encontra 682 imagens distintas e ~100 MB de cópias.
//
// O índice guarda as duas leituras que interessam:
//
//   - ORIGINAL (data/): o grupo "sempre foi igual" — é ele que define o
//     alvo da propagação de import (as cópias de DVD voltam a ficar
//     sincronizadas juntas). O pareamento é SEMPRE de data/: em mods/
//     qualquer uma das cópias pode estar editada, então olhar mods/ para
//     agrupar misturaria original com reimpressão. Id sem original em
//     data/ (só em mods/, ou a árvore inteira em modo fallback) fica FORA
//     do índice — nesses casos não há grupo confiável para agrupar.
//   - EFETIVA (mods-first): o que o jogo realmente carrega — mostra quais
//     cópias ainda estão idênticas depois de um import.
//
// O índice é por VERSÃO: os grupos nunca cruzam FFX ↔ FFX-2.
//
// Não há hard link nem dedupe em disco: o file loader do jogo e a cópia da
// árvore para a library do Steam tornariam links frágeis. O índice é só
// informação para o translator.

// PayloadHash devolve o SHA-256 (hex) dos dados da imagem e o tamanho do
// payload em bytes. Erro = container ilegível (o chamador pula o arquivo).
func PayloadHash(raw []byte) (hash string, payloadSize int64, err error) {
	off, dataErr := dataOffsetOf(raw)
	if dataErr != nil {
		return "", 0, dataErr
	}
	sum := sha256.Sum256(raw[off:])
	return hex.EncodeToString(sum[:]), int64(len(raw)) - int64(off), nil
}

// Copy é uma cópia de mesma imagem, com o estado relativo à textura
// consultada.
type Copy struct {
	// ID é o id da cópia (caminho sem sufixo .dds.phyre).
	ID string
	// InMods indica que a cópia tem substituição em mods/ (já importada).
	InMods bool
	// Identical indica que o payload efetivo (mods-first, o que o jogo
	// carrega) é o mesmo da textura consultada — cópia ainda sincronizada.
	Identical bool
}

// entry é o que o índice sabe de uma textura.
type entry struct {
	origHash string
	effHash  string
	payload  int64
	inMods   bool
}

// Index agrupa as texturas de uma versão pelo hash do payload.
type Index struct {
	byID   map[string]entry
	byOrig map[string][]string
	byEff  map[string][]string
}

// BuildIndex varre a árvore da versão e monta o índice do zero (sem cache).
//
// Só entram ids com original em data/ — o pareamento é decidido em data/,
// como manda a regra do projeto; em mods/ qualquer cópia pode estar
// editada, então agrupar por mods/ misturaria pristine com reimpressão. É
// também o que desliga o recurso no modo fallback (data/ vazio: índice
// vazio, árvore completa visível) e tira daqui ids só-mods, que a regra 4
// nem mostra na árvore.
func BuildIndex(version common.GameVersion) (*Index, error) {
	ids, _, _, err := Scan(version)
	if err != nil {
		return nil, err
	}
	ix := &Index{
		byID:   make(map[string]entry, len(ids)),
		byOrig: make(map[string][]string),
		byEff:  make(map[string][]string),
	}
	for _, id := range ids {
		rel := RelPath(version, id)

		origRaw, origErr := readSource(rel, common.SourceData)
		if origErr != nil {
			// Sem pristine em data/: este id não participa de grupo algum.
			continue
		}
		origHash, payload, perr := PayloadHash(origRaw)
		if perr != nil {
			common.LogWarning("images %s/%s: payload ilegível (fora do índice): %v", version, id, perr)
			continue
		}

		effHash, inMods := origHash, false
		if modsRaw, merr := readSource(rel, common.SourceMods); merr == nil {
			if h, _, herr := PayloadHash(modsRaw); herr == nil {
				effHash, inMods = h, true
			} else {
				common.LogWarning("images %s/%s: mods/ ilegível, usando o original: %v", version, id, herr)
			}
		}

		ix.byID[id] = entry{origHash: origHash, effHash: effHash, payload: payload, inMods: inMods}
		ix.byOrig[origHash] = append(ix.byOrig[origHash], id)
		ix.byEff[effHash] = append(ix.byEff[effHash], id)
	}
	for _, group := range ix.byOrig {
		sort.Strings(group)
	}
	for _, group := range ix.byEff {
		sort.Strings(group)
	}
	return ix, nil
}

// Copies devolve as OUTRAS texturas de mesmo payload original que id (em
// ordem canônica) e o tamanho do payload, para o frontend mostrar o
// desperdício do grupo (payload × (N-1)).
func (ix *Index) Copies(id string) (copies []Copy, payload int64) {
	if ix == nil {
		return nil, 0
	}
	self, ok := ix.byID[id]
	if !ok {
		return nil, 0
	}
	group := ix.byOrig[self.origHash]
	if len(group) < 2 {
		return nil, self.payload
	}
	copies = make([]Copy, 0, len(group)-1)
	for _, other := range group {
		if other == id {
			continue
		}
		e := ix.byID[other]
		copies = append(copies, Copy{
			ID:        other,
			InMods:    e.inMods,
			Identical: e.effHash == self.effHash,
		})
	}
	return copies, self.payload
}

// OriginalGroup devolve o grupo de mesmo payload original (inclui id, em
// ordem canônica) — é ele que valida a propagação de import: qualquer alvo
// fora do grupo é recusado antes de gravar um byte.
func (ix *Index) OriginalGroup(id string) []string {
	if ix == nil {
		return nil
	}
	self, ok := ix.byID[id]
	if !ok {
		return nil
	}
	return ix.byOrig[self.origHash]
}

// IndexFor devolve o índice da versão, construindo na primeira chamada e
// memorizando depois (a varredura lê a árvore inteira, ~1-2 s).
//
// O cache é invalidado por InvalidateIndex (import muda o payload efetivo);
// edição externa no hex editor só o "Reanalisar" do painel pega.
func IndexFor(version common.GameVersion) (*Index, error) {
	indexMu.Lock()
	defer indexMu.Unlock()
	if ix, ok := indexCache[version]; ok {
		return ix, nil
	}
	ix, err := BuildIndex(version)
	if err != nil {
		return nil, err
	}
	indexCache[version] = ix
	return ix, nil
}

// InvalidateIndex descarta o índice em cache da versão (próxima consulta
// reconstrói).
func InvalidateIndex(version common.GameVersion) {
	indexMu.Lock()
	defer indexMu.Unlock()
	delete(indexCache, version)
}

// InvalidateIndexes descarta todos os índices em cache.
func InvalidateIndexes() {
	indexMu.Lock()
	defer indexMu.Unlock()
	indexCache = map[common.GameVersion]*Index{}
}

var (
	indexMu    sync.Mutex
	indexCache = map[common.GameVersion]*Index{}
)

// readSource lê o .dds.phyre da fonte dada (data/ ou mods/); erro quando o
// arquivo não existe ou não pode ser lido.
func readSource(rel string, source common.FileSource) ([]byte, error) {
	acc, err := common.NewFileAccessorFrom(rel, source)
	if err != nil {
		return nil, err
	}
	if !acc.Exists {
		return nil, errMissing
	}
	return acc.ReadBytes()
}
