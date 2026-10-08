// Package vbf lê o container Virtuos Big File (.vbf) dos remasters de
// FFX/FFX-2. SOMENTE LEITURA: o pacote nunca abre nada em escrita e nunca
// materializa a árvore inteira — só o cabeçalho (índice) e os blocos do
// arquivo pedido, sob demanda.
//
// O layout abaixo foi reconstruído empiricamente: um formato de autoria
// desconhecida, decompilado de uma ferramenta que ninguém sabe quem fez, e
// o leitor de .vbf embutido no próprio launcher do jogo — conferidos contra
// FFX_Data.vbf e FFX2_Data.vbf reais (assinatura, índice, blocos e MD5 do
// cabeçalho batem).
//
// Layout (little-endian):
//
//	u32 magic "SRYK" (0x4B595253)
//	u32 headerLen          — bytes do cabeçalho, do offset 0 ao fim da lista
//	                         de blocos (o que é coberto pelo MD5 trailer)
//	u64 numFiles
//	numFiles × [16]MD5(path)      — hash do caminho (conferência/lookup)
//	numFiles × 32:  startBlock u32, 0 u32, bytes u64, offset u64, nameOffset u64
//	u32 nameTableLen (inclui os próprios 4 bytes) + nameTable
//	blocks × u16             — tamanho compactado de cada bloco (0 = 64 KiB)
//	…payload dos arquivos…
//	[16]MD5(header[0:headerLen]) — últimos 16 bytes do arquivo
//
// O payload de um arquivo é uma sequência de blocos de 65536 bytes de
// saída. Cada bloco é CRU quando o u16 vale 0 (mapeado para 65536) e
// deflate RAW caso contrário (os 2 bytes de cabeçalho zlib são pulados —
// é assim que o leitor de referência descompacta).
package vbf

import (
	"bufio"
	"bytes"
	"compress/flate"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// Magic é a assinatura "SRYK" lida como u32 little-endian.
	Magic uint32 = 0x4B595253

	// blockSize é o tamanho (de SAÍDA) de cada bloco comprimido.
	blockSize = 64 << 10

	// MaxDecodeBytes limita a decodificação de UM arquivo em memória: os
	// .vbf guardam vídeos e executáveis de centenas de MB que ninguém abre
	// na tabela nem no painel de imagem.
	MaxDecodeBytes = 64 << 20
)

// Erros estáveis do pacote (o frontend os repassa como mensagem).
var (
	// ErrNotVBF é devolvido quando a assinatura não é "SRYK".
	ErrNotVBF = errors.New("vbf: assinatura inválida (não é um .vbf)")
	// ErrCorrupt é devolvido quando o cabeçalho não bate com o layout.
	ErrCorrupt = errors.New("vbf: cabeçalho corrompido")
	// ErrNotFound é devolvido quando o caminho não está no índice.
	ErrNotFound = errors.New("vbf: arquivo não encontrado")
	// ErrTooLarge é devolvido quando o arquivo pedido excede MaxDecodeBytes.
	ErrTooLarge = errors.New("vbf: arquivo grande demais para decodificar em memória")
)

// Entry é a linha do índice de um arquivo do container.
type Entry struct {
	// Path é o caminho dentro do container, com "/" (ex.:
	// "ffx_ps2/ffx/master/new_uspc/menu/macrodic.dcp").
	Path string
	// Size é o tamanho DESCOMPACTIDO do arquivo.
	Size uint64
	// StartBlock é o índice do 1º bloco do arquivo na lista de blocos.
	StartBlock uint32
	// DataOffset é o offset ABSOLUTO no .vbf do 1º bloco do arquivo.
	DataOffset uint64
}

// Node é um filho de diretório devolvido por Archive.List.
type Node struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  uint64 `json:"size,omitempty"`
}

// Archive é um .vbf aberto: índice em memória + handle de leitura.
// ReadAt é seguro para concorrência; Close deve ser chamado uma vez.
type Archive struct {
	path     string
	file     *os.File
	size     int64
	entries  []Entry
	index    map[string]int // chave: caminho normalizado (minúsculo, "/")
	dirs     map[string]struct{}
	blocks   []uint16
	numFiles int
}

// normaliza devolve a chave de índice/caminho: "/" como separador, sem
// "./", sem barra inicial e em minúsculas (o próprio índice guarda o MD5
// do caminho em minúsculas — o lookup do formato é case-insensitive).
func normaliza(p string) string {
	s := strings.ReplaceAll(p, "\\", "/")
	s = path.Clean("/" + s)
	s = strings.TrimPrefix(s, "/")
	if s == "." {
		s = ""
	}
	return strings.ToLower(s)
}

// u64le lê um u64 little-endian de 8 bytes.
func u64le(b []byte) uint64 {
	return uint64(binary.LittleEndian.Uint32(b[0:4])) |
		uint64(binary.LittleEndian.Uint32(b[4:8]))<<32
}

// blocksFor devolve quantos blocos um arquivo de size bytes ocupa.
func blocksFor(size uint64) int {
	if size == 0 {
		return 0
	}
	return int((size + blockSize - 1) / blockSize)
}

// Open abre o .vbf, valida a assinatura e o MD5 do cabeçalho e carrega o
// índice (entries + nomes + lista de blocos) em memória. O payload NÃO é
// lido aqui: um .vbf de 20 GB só custa o cabeçalho (~10-18 MB).
func Open(filePath string) (*Archive, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			_ = f.Close()
		}
	}()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if st.Size() < 16+md5.Size {
		return nil, fmt.Errorf("%w: arquivo pequeno demais (%d bytes)", ErrCorrupt, st.Size())
	}

	var head [16]byte
	if _, err := io.ReadFull(f, head[:]); err != nil {
		return nil, err
	}
	if binary.LittleEndian.Uint32(head[0:4]) != Magic {
		return nil, ErrNotVBF
	}
	headerLen := uint64(binary.LittleEndian.Uint32(head[4:8]))
	numFiles := binary.LittleEndian.Uint64(head[8:16])

	// Teto de sanidade: cada arquivo ocupa 48 bytes de índice, além dos
	// 16 iniciais e dos 16 do trailer.
	if headerLen < 16 || headerLen > uint64(st.Size())-md5.Size {
		return nil, fmt.Errorf("%w: headerLen=%d incompatível com arquivo de %d bytes", ErrCorrupt, headerLen, st.Size())
	}
	if numFiles > (headerLen-16)/48 {
		return nil, fmt.Errorf("%w: numFiles=%d não cabe no cabeçalho", ErrCorrupt, numFiles)
	}

	a := &Archive{
		path:     filePath,
		file:     f,
		size:     st.Size(),
		numFiles: int(numFiles),
		entries:  make([]Entry, 0, numFiles),
		index:    make(map[string]int, numFiles),
		dirs:     make(map[string]struct{}),
	}

	// Leitor do corpo do cabeçalho, a partir do offset 16 (o head já lido).
	rest := io.NewSectionReader(f, 16, int64(headerLen)-16)
	r := bufferedReader(rest)

	// 1. hashes MD5 dos caminhos: só pulamos (o índice é pelo nome).
	if _, err := io.CopyN(io.Discard, r, int64(numFiles)*16); err != nil {
		return nil, fmt.Errorf("%w: lendo hashes de caminho: %v", ErrCorrupt, err)
	}

	// 2. entradas do índice.
	type rawEntry struct {
		startBlock uint32
		size       uint64
		offset     uint64
		nameOffset uint64
	}
	raws := make([]rawEntry, 0, numFiles)
	sizes := make([]uint64, 0, numFiles)
	var scratch [32]byte
	for i := uint64(0); i < numFiles; i++ {
		if _, err := io.ReadFull(r, scratch[:]); err != nil {
			return nil, fmt.Errorf("%w: lendo entry %d: %v", ErrCorrupt, i, err)
		}
		re := rawEntry{
			startBlock: binary.LittleEndian.Uint32(scratch[0:4]),
			size:       u64le(scratch[8:16]),
			offset:     u64le(scratch[16:24]),
			nameOffset: u64le(scratch[24:32]),
		}
		// O campo scratch[4:8] é zero no formato.
		raws = append(raws, re)
		sizes = append(sizes, re.size)
	}

	// 3. tabela de nomes: u32 inclui os próprios 4 bytes.
	var nameSizeBuf [4]byte
	if _, err := io.ReadFull(r, nameSizeBuf[:]); err != nil {
		return nil, fmt.Errorf("%w: lendo tamanho da tabela de nomes: %v", ErrCorrupt, err)
	}
	nameTableLen := binary.LittleEndian.Uint32(nameSizeBuf[:])
	if nameTableLen < 4 {
		return nil, fmt.Errorf("%w: tabela de nomes com %d bytes", ErrCorrupt, nameTableLen)
	}
	names := make([]byte, nameTableLen-4)
	if _, err := io.ReadFull(r, names); err != nil {
		return nil, fmt.Errorf("%w: lendo tabela de nomes: %v", ErrCorrupt, err)
	}

	// 4. lista de blocos: contagem derivada das próprias entries.
	blockCount := 0
	for _, sz := range sizes {
		blockCount += blocksFor(sz)
	}
	a.blocks = make([]uint16, blockCount)
	var b2 [2]byte
	for i := 0; i < blockCount; i++ {
		if _, err := io.ReadFull(r, b2[:]); err != nil {
			return nil, fmt.Errorf("%w: lendo lista de blocos: %v", ErrCorrupt, err)
		}
		a.blocks[i] = binary.LittleEndian.Uint16(b2[:])
	}

	// Conferência de FECHAMENTO: o que se leu tem de bater exatamente com
	// headerLen (senão a estrutura acima está errada). O campo de tamanho da
	// tabela de nomes JÁ está contido em nameTableLen (inclui os 4 bytes).
	expected := 16 + uint64(numFiles)*48 + uint64(nameTableLen) + uint64(blockCount)*2
	if expected != headerLen {
		return nil, fmt.Errorf("%w: layout fecha em %d mas headerLen=%d", ErrCorrupt, expected, headerLen)
	}

	// 5. nomes + índice.
	for i, re := range raws {
		if re.nameOffset >= uint64(len(names)) {
			return nil, fmt.Errorf("%w: nameOffset %d da entry %d fora da tabela (%d)",
				ErrCorrupt, re.nameOffset, i, len(names))
		}
		n := names[re.nameOffset:]
		end := bytes.IndexByte(n, 0)
		if end < 0 {
			end = len(n)
		}
		name := strings.TrimSpace(string(n[:end]))
		if name == "" {
			return nil, fmt.Errorf("%w: entry %d sem nome", ErrCorrupt, i)
		}
		e := Entry{
			Path:       strings.ReplaceAll(name, "\\", "/"),
			Size:       re.size,
			StartBlock: re.startBlock,
			DataOffset: re.offset,
		}
		key := normaliza(e.Path)
		if _, dup := a.index[key]; !dup {
			a.index[key] = i
		}
		a.entries = append(a.entries, e)
		for slash := strings.IndexByte(key, '/'); slash >= 0; {
			a.dirs[key[:slash]] = struct{}{}
			next := strings.IndexByte(key[slash+1:], '/')
			if next < 0 {
				break
			}
			slash += next + 1
		}
	}

	// 6. MD5 do cabeçalho vs. os 16 últimos bytes do arquivo.
	sum := md5.New()
	if _, err := io.Copy(sum, io.NewSectionReader(f, 0, int64(headerLen))); err != nil {
		return nil, fmt.Errorf("%w: lendo cabeçalho para o MD5: %v", ErrCorrupt, err)
	}
	trailer := make([]byte, md5.Size)
	if _, err := f.ReadAt(trailer, st.Size()-int64(md5.Size)); err != nil {
		return nil, fmt.Errorf("%w: lendo trailer MD5: %v", ErrCorrupt, err)
	}
	if !bytes.Equal(sum.Sum(nil), trailer) {
		return nil, fmt.Errorf("%w: MD5 do cabeçalho não confere", ErrCorrupt)
	}

	ok = true
	return a, nil
}

// bufferedReader devolve um reader com buffer para o seccionador do
// cabeçalho (milhares de ReadFull pequenos sem uma syscall a cada um).
func bufferedReader(r io.Reader) io.Reader {
	return bufio.NewReaderSize(r, 64<<10)
}

// Close libera o handle do arquivo.
func (a *Archive) Close() error {
	if a == nil || a.file == nil {
		return nil
	}
	f := a.file
	a.file = nil
	return f.Close()
}

// Path devolve o caminho do .vbf no disco.
func (a *Archive) Path() string { return a.path }

// Len devolve a quantidade de arquivos no índice.
func (a *Archive) Len() int { return a.numFiles }

// Size devolve o tamanho do .vbf em disco.
func (a *Archive) Size() int64 { return a.size }

// Entry resolve um caminho do índice (case-insensitive, "/" como
// separador). ok=false quando o caminho não existe.
func (a *Archive) Entry(name string) (Entry, bool) {
	if a == nil {
		return Entry{}, false
	}
	i, ok := a.index[normaliza(name)]
	if !ok {
		return Entry{}, false
	}
	return a.entries[i], true
}

// Has confere a presença de um caminho no índice.
func (a *Archive) Has(name string) bool {
	_, ok := a.Entry(name)
	return ok
}

// IsDir informa se o caminho tem filhos no índice (mesmo se existir uma
// entry de arquivo com o mesmo caminho; List trata o diretório como vencedor).
func (a *Archive) IsDir(dir string) bool {
	if a == nil {
		return false
	}
	d := normaliza(dir)
	if d == "" {
		return len(a.entries) > 0
	}
	_, ok := a.dirs[d]
	return ok
}

// Paths devolve todos os caminhos do índice (diagnóstico/testes).
func (a *Archive) Paths() []string {
	out := make([]string, 0, len(a.entries))
	for _, e := range a.entries {
		out = append(out, e.Path)
	}
	return out
}

// List devolve os filhos IMEDIATOS de dir ("" = raiz do container).
// Diretórios vêm antes dos arquivos, ambos em ordem alfabética.
func (a *Archive) List(dir string) []Node {
	prefix := ""
	if d := normaliza(dir); d != "" {
		prefix = d + "/"
	}

	dirs := map[string]string{} // chave normalizada → caminho original
	files := map[string]Node{}
	for _, e := range a.entries {
		kp := normaliza(e.Path)
		if !strings.HasPrefix(kp, prefix) {
			continue
		}
		rest := kp[len(prefix):]
		if rest == "" {
			// É o próprio dir pedido.
			continue
		}
		// Sufixo original (mesmo comprimento: a normalização só troca o
		// separador e a caixa), para exibir o nome como está no .vbf.
		origRest := e.Path[len(e.Path)-len(rest):]
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			dirs[prefix+rest[:i]] = prefix + origRest[:i]
			continue
		}
		files[prefix+rest] = Node{
			Name:  origRest,
			Path:  e.Path,
			IsDir: false,
			Size:  e.Size,
		}
	}

	// Um caminho pode ser arquivo E prefixo de diretório: o diretório vence
	// (é ele que a árvore precisa expandir).
	out := make([]Node, 0, len(dirs)+len(files))
	for key, orig := range dirs {
		delete(files, key)
		out = append(out, Node{
			Name:  orig[strings.LastIndex(orig, "/")+1:],
			Path:  orig,
			IsDir: true,
		})
	}
	for _, n := range files {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].IsDir != out[j].IsDir {
			return out[i].IsDir
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// Read decodifica o arquivo do caminho pedido (índice → blocos) e devolve
// os bytes DESCOMPACTADOS em memória. Caminho desconhecido → ErrNotFound;
// arquivo maior que MaxDecodeBytes → ErrTooLarge (nunca tenta ler 20 GB).
func (a *Archive) Read(name string) ([]byte, error) {
	e, ok := a.Entry(name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return a.ReadEntry(e)
}

// FilesUnder devolve as entries do índice cujo caminho está DENTRO de dir
// ("" = raiz = todas as entries). A ordem é a do índice (caminho do
// container). É a expansão da seleção da árvore: um diretório marcado vira
// todos os arquivos abaixo dele.
func (a *Archive) FilesUnder(dir string) []Entry {
	prefix := ""
	if d := normaliza(dir); d != "" {
		prefix = d + "/"
	}
	out := make([]Entry, 0, len(a.entries))
	for _, e := range a.entries {
		if prefix == "" || strings.HasPrefix(normaliza(e.Path), prefix) {
			out = append(out, e)
		}
	}
	return out
}

// ExtractEntry decodifica a entry (índice → blocos) gravando direto em
// destPath, bloco a bloco — sem carregar o arquivo em memória e SEM o
// limite MaxDecodeBytes (é o caminho da EXTRAÇÃO: vídeos e executáveis de
// centenas de MB saem inteiros). O arquivo é gravado em .tmp e renomeado
// no fim: uma extração interrompida nunca deixa meio arquivo com cara de
// pronto.
func (a *Archive) ExtractEntry(e Entry, destPath string) error {
	if a == nil || a.file == nil {
		return fmt.Errorf("%w: container fechado", ErrCorrupt)
	}
	count := blocksFor(e.Size)
	remainder := e.Size % blockSize
	if remainder == 0 {
		remainder = blockSize
	}
	if int(e.StartBlock)+count > len(a.blocks) {
		return fmt.Errorf("%w: %s aponta para bloco fora do índice", ErrCorrupt, e.Path)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("criando diretório de %s: %w", e.Path, err)
	}
	f, err := os.CreateTemp(filepath.Dir(destPath), ".vbf-extract-*.tmp")
	if err != nil {
		return fmt.Errorf("criando %s: %w", destPath, err)
	}
	tmp := f.Name()
	ok := false
	defer func() {
		if !ok {
			_ = f.Close()
			_ = os.Remove(tmp)
		}
	}()
	if err := f.Chmod(0o644); err != nil {
		return fmt.Errorf("ajustando permissões de %s: %w", destPath, err)
	}

	w := bufio.NewWriterSize(f, blockSize)
	off := int64(e.DataOffset)
	for i := 0; i < count; i++ {
		stored := int(a.blocks[int(e.StartBlock)+i])
		if stored == 0 {
			stored = blockSize
		}
		if stored < 0 || off < 0 || off+int64(stored) > a.size {
			return fmt.Errorf("%w: %s: bloco %d fora do arquivo", ErrCorrupt, e.Path, i)
		}
		buf := make([]byte, stored)
		if _, err := a.file.ReadAt(buf, off); err != nil {
			return fmt.Errorf("%w: %s: lendo bloco %d: %v", ErrCorrupt, e.Path, i, err)
		}
		off += int64(stored)

		want := blockSize
		last := i == count-1
		if last {
			want = int(remainder)
		}
		// Mesma regra do ReadEntry: bloco cru quando o u16 vale o tamanho
		// de bloco cheio (ou o exato resto no último bloco).
		if stored == blockSize || (last && stored == want) {
			if len(buf) < want {
				return fmt.Errorf("%w: %s: bloco %d trucado", ErrCorrupt, e.Path, i)
			}
			if _, err := w.Write(buf[:want]); err != nil {
				return fmt.Errorf("gravando %s: %w", destPath, err)
			}
			continue
		}
		if len(buf) < 2 {
			return fmt.Errorf("%w: %s: bloco %d compactado inválido", ErrCorrupt, e.Path, i)
		}
		dec, err := inflateExact(buf[2:], want)
		if err != nil {
			return fmt.Errorf("%w: %s: descompactando bloco %d: %v", ErrCorrupt, e.Path, i, err)
		}
		if _, err := w.Write(dec); err != nil {
			return fmt.Errorf("gravando %s: %w", destPath, err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("gravando %s: %w", destPath, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("fechando %s: %w", destPath, err)
	}
	if info, err := os.Lstat(destPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || info.IsDir() {
			return fmt.Errorf("destino não é um arquivo regular: %s", destPath)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("verificando destino %s: %w", destPath, err)
	}
	if err := os.Rename(tmp, destPath); err != nil {
		return fmt.Errorf("renomeando para %s: %w", destPath, err)
	}
	ok = true
	return nil
}

// ReadEntry decodifica uma entry do índice.
func (a *Archive) ReadEntry(e Entry) ([]byte, error) {
	if e.Size == 0 {
		return []byte{}, nil
	}
	if e.Size > MaxDecodeBytes {
		return nil, fmt.Errorf("%w: %s tem %d bytes (limite %d)",
			ErrTooLarge, e.Path, e.Size, MaxDecodeBytes)
	}
	count := blocksFor(e.Size)
	remainder := e.Size % blockSize
	if remainder == 0 {
		remainder = blockSize
	}
	if int(e.StartBlock)+count > len(a.blocks) {
		return nil, fmt.Errorf("%w: %s aponta para bloco fora do índice", ErrCorrupt, e.Path)
	}

	out := make([]byte, 0, e.Size)
	off := int64(e.DataOffset)
	for i := 0; i < count; i++ {
		stored := int(a.blocks[int(e.StartBlock)+i])
		if stored == 0 {
			stored = blockSize
		}
		if stored < 0 || off < 0 || off+int64(stored) > a.size {
			return nil, fmt.Errorf("%w: %s: bloco %d fora do arquivo", ErrCorrupt, e.Path, i)
		}
		buf := make([]byte, stored)
		if _, err := a.file.ReadAt(buf, off); err != nil {
			return nil, fmt.Errorf("%w: %s: lendo bloco %d: %v", ErrCorrupt, e.Path, i, err)
		}
		off += int64(stored)

		want := blockSize
		last := i == count-1
		if last {
			want = int(remainder)
		}
		// stored==65536 (vindo do 0) é sempre cru; no último bloco, um
		// tamanho igual ao que vai ser escrito também é cru — é a regra
		// que o leitor de referência usa para distinguir bloco cru de
		// bloco compactado.
		if stored == blockSize || (last && stored == want) {
			if len(buf) < want {
				return nil, fmt.Errorf("%w: %s: bloco %d trucado", ErrCorrupt, e.Path, i)
			}
			out = append(out, buf[:want]...)
			continue
		}
		if len(buf) < 2 {
			return nil, fmt.Errorf("%w: %s: bloco %d compactado inválido", ErrCorrupt, e.Path, i)
		}
		dec, err := inflateExact(buf[2:], want)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: descompactando bloco %d: %v", ErrCorrupt, e.Path, i, err)
		}
		out = append(out, dec...)
	}
	return out, nil
}

// inflateExact descompacta deflate RAW (cabeçalho zlib de 2 bytes já
// removido pelo chamador) exigindo exatamente `want` bytes de saída —
// qualquer desvio é corrupção, não "melhor esforço".
func inflateExact(b []byte, want int) ([]byte, error) {
	fr := flate.NewReader(bytes.NewReader(b))
	defer fr.Close()
	data, err := io.ReadAll(io.LimitReader(fr, int64(want)+1))
	if err != nil {
		return nil, err
	}
	if len(data) != want {
		return nil, fmt.Errorf("saída com %d bytes (esperado %d)", len(data), want)
	}
	return data, nil
}
