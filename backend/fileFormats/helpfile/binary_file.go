// Package helpfile implementa o formato .sps2 do Painel de ajuda do FFX
// (help/now_help/now_help.sps2 e now_help_page.sps2). Formato nativo do
// projeto: nada de shsplit/arquivos legados.
//
// Layout (todos os campos de header são uint32 little-endian):
//
//	0x00 magic (1)
//	0x04 pageCount (0 nos arquivos simples; N nos *_page — estático)
//	0x08 textEnd (fim do bloco de texto — recalculado no rebuild)
//	0x0C ptrBase (0x24 quando há sub-header de páginas; 0 caso contrário)
//	0x10 ptrTableOff (início da tabela de ponteiros; 0x24 ou 0x114)
//	0x14 footerPtr (início da árvore/footer do bloco de páginas; 0 se ausente)
//	0x18..0x24 timestamps (12 bytes preservados verbatim)
//	[ptrBase..ptrTableOff]      sub-header de páginas (preservado verbatim)
//	[ptrTableOff..+4N]          tabela de N ponteiros u32 (N = (ptr[0]-ptrTableOff)/4)
//	[ptr[0]..textEnd]           segmentos: texto + 0x00 (+ bloco de dados opcional)
//	[textEnd..footerPtr]        registros de páginas {u32 offset; u16 size; u16 FFFF}
//	                            + dados das páginas (preservados; offsets deslocados)
//	[footerPtr..EOF]            árvore/footer (preservado verbatim)
//
// Todo segmento contém o terminador 0x00 (padrão do formato): o texto vai até
// o 0x00; o que sobra no segmento (antes do próximo ponteiro) é o bloco de
// dados, preservado byte a byte no rebuild.
package helpfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
)

const (
	// helpFixedHeaderLen é o tamanho fixo do header: 6 dwords + 12 bytes
	// de timestamps (0x18..0x24).
	helpFixedHeaderLen = 0x24

	// HelpMagic é o valor do dword inicial do arquivo (assinatura).
	HelpMagic uint32 = 0x00000001

	// HelpRecordSeparator fecha cada registro do bloco de páginas.
	HelpRecordSeparator uint16 = 0xFFFF
)

// Nomes válidos de arquivo dos painéis de ajuda (kind help). Cada um vive no
// subdiretório de mesmo nome dentro de help/ (exceto now_help_page, irmão de
// now_help). Todos FFX-only: a árvore ffx2 não tem a pasta help/.
const (
	NowHelpName     = "now_help"
	NowHelpPageName = "now_help_page"
	MonBokuName     = "mon_boku"
	SMonitorName    = "s_monitor"
	DvdcopyName     = "dvdcopy"
	DvdcopyPageName = "dvdcopy_page"
	HelpFileExt     = ".sps2"
	HelpDirPrefix   = "help"
)

// HelpEntry mapeia o id (stem do arquivo) ao subdiretório de help/.
type HelpEntry struct {
	Name string
	Dir  string
}

// HelpEntries é o registro dos painéis em ordem determinística (o
// ListEntries do backend e a sidebar do frontend seguem esta ordem).
var HelpEntries = []HelpEntry{
	{Name: NowHelpName, Dir: NowHelpName},
	{Name: NowHelpPageName, Dir: NowHelpName},
	{Name: MonBokuName, Dir: MonBokuName},
	{Name: SMonitorName, Dir: SMonitorName},
	{Name: DvdcopyName, Dir: DvdcopyName},
	{Name: DvdcopyPageName, Dir: DvdcopyName},
}

// helpEntryByName indexa HelpEntries por id.
var helpEntryByName = func() map[string]HelpEntry {
	m := make(map[string]HelpEntry, len(HelpEntries))
	for _, e := range HelpEntries {
		m[e.Name] = e
	}
	return m
}()

// IsHelpEntry informa se o id é um painel conhecido do formato.
func IsHelpEntry(name string) bool {
	_, ok := helpEntryByName[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

// HelpEntryNames devolve os ids em ordem de registro.
func HelpEntryNames() []string {
	out := make([]string, 0, len(HelpEntries))
	for _, e := range HelpEntries {
		out = append(out, e.Name)
	}
	return out
}

// HelpEntryDir devolve o subdiretório do painel ("" se desconhecido).
func HelpEntryDir(name string) string {
	if e, ok := helpEntryByName[strings.ToLower(strings.TrimSpace(name))]; ok {
		return e.Dir
	}
	return ""
}

// HelpRelPath devolve o caminho relativo (dentro da raiz de localização) do
// arquivo .sps2 do painel: help/<dir>/<name>.sps2.
func HelpRelPath(name string) string {
	dir := HelpEntryDir(name)
	if dir == "" {
		dir = NowHelpName
	}
	return filepath.Join(HelpDirPrefix, dir, name+HelpFileExt)
}

// HelpPathForVersion resolve o caminho do arquivo para versão/localização.
func HelpPathForVersion(version common.GameVersion, localization, name string) string {
	return filepath.Join(
		common.GetLocalizationRootForVersion(version, localization),
		HelpRelPath(name),
	)
}

// HelpModsPathForVersion resolve o caminho de GRAVAÇÃO no padrão mods/:
// mods/<ffx_ps2/<versão>/master/new_<loc>pc>/help/<dir>/<name>.sps2.
// É o mesmo padrão dos demais formatos (o original do gamefiles fica intacto).
func HelpModsPathForVersion(version common.GameVersion, localization, name string) string {
	return filepath.Join(common.GameFilesRoot, common.ModsFolder, HelpPathForVersion(version, localization, name))
}

// HelpSegment é um segmento de texto do arquivo: os ponteiros delimitam os
// segmentos (o limite do segmento i é ptr[i+1]; o último fecha em textEnd).
//
// Estrutura confirmada: todo segmento é texto + terminador 0x00 + bloco de
// dados opcional (presente em 15/16 dos 1208 segmentos).
type HelpSegment struct {
	Index int

	// RawBytes são os bytes originais do texto até e incluindo o 0x00.
	// Usados verbatim quando o texto não foi alterado — garante round-trip
	// byte-exato independentemente da fidelidade decode→encode do conversor.
	RawBytes []byte

	// DataBytes são os bytes do segmento após o 0x00 (bloco de dados
	// preservado verbatim; vazio na maioria dos segmentos).
	DataBytes []byte

	// Text é o texto decodificado (charset da localização).
	Text string

	// dirty marca que Text foi alterado e o rebuild deve reencodar.
	dirty bool
}

// SetText altera o texto do segmento (o rebuild reencodará via converter).
func (s *HelpSegment) SetText(text string) {
	if text != s.Text {
		s.Text = text
		s.dirty = true
	}
}

// HelpBinaryFile é o arquivo .sps2 carregado. Todas as partes ficam em
// slices para a reconstrução: header reescrito, ponteiros recalculados e
// blocos estruturais (sub-header, páginas, footer) preservados.
type HelpBinaryFile struct {
	Version      common.GameVersion
	Localization string
	Name         string

	Magic       uint32
	PageCount   uint32
	TextEnd     uint32
	PtrBase     uint32
	PtrTableOff uint32
	FooterPtr   uint32
	TimeStamp   [helpFixedHeaderLen - 0x18]byte

	SubHeader  []byte         // ptrBase..ptrTableOff (nil nos arquivos simples)
	Pointers   []uint32       // N entradas
	Segments   []*HelpSegment // N segmentos
	PagesBlock []byte         // textEnd..footerPtr (só *_page)
	Footer     []byte         // footerPtr..EOF (só *_page)

	// SourcePath é o caminho absoluto resolvido de onde o binário foi lido
	// (mods-first). A gravação (Save) vai para a árvore mods/.
	SourcePath string
}

// Charset devolve o charset de encode da localização do arquivo.
func (f *HelpBinaryFile) Charset() string {
	if f.Localization == "" {
		return common.DefaultLocalization
	}
	return f.Localization
}

// IsDirty informa se algum segmento foi alterado desde o parse.
func (f *HelpBinaryFile) IsDirty() bool {
	for _, seg := range f.Segments {
		if seg.dirty {
			return true
		}
	}
	return false
}

func readUint32(data []byte, off int) (uint32, error) {
	if off < 0 || off+4 > len(data) {
		return 0, fmt.Errorf("helpfile: offset 0x%X fora do arquivo (len=0x%X)", off, len(data))
	}
	return binary.LittleEndian.Uint32(data[off:]), nil
}

func cloneBytes(b []byte) []byte {
	if len(b) == 0 {
		return nil
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out
}

// ReadHelpBinary parseia o binário .sps2 em memória.
func ReadHelpBinary(data []byte, name, localization string, version common.GameVersion) (*HelpBinaryFile, error) {
	if len(data) < helpFixedHeaderLen {
		return nil, fmt.Errorf("helpfile: %s muito pequeno para o header (%d bytes)", name, len(data))
	}

	magic := binary.LittleEndian.Uint32(data[0x00:])
	if magic != HelpMagic {
		return nil, fmt.Errorf("helpfile: %s magic inválido: 0x%08X (esperado 0x%08X)", name, magic, HelpMagic)
	}

	f := &HelpBinaryFile{
		Version:      version,
		Localization: localization,
		Name:         name,
		Magic:        magic,
		PageCount:    binary.LittleEndian.Uint32(data[0x04:]),
		TextEnd:      binary.LittleEndian.Uint32(data[0x08:]),
		PtrBase:      binary.LittleEndian.Uint32(data[0x0C:]),
		PtrTableOff:  binary.LittleEndian.Uint32(data[0x10:]),
		FooterPtr:    binary.LittleEndian.Uint32(data[0x14:]),
	}
	copy(f.TimeStamp[:], data[0x18:helpFixedHeaderLen])

	if int(f.PtrTableOff)+4 > len(data) {
		return nil, fmt.Errorf("helpfile: %s ptrTableOff 0x%X fora do arquivo", name, f.PtrTableOff)
	}
	if f.PtrTableOff < helpFixedHeaderLen {
		return nil, fmt.Errorf("helpfile: %s ptrTableOff 0x%X menor que o header fixo", name, f.PtrTableOff)
	}
	if f.PtrBase != 0 && f.PtrBase != helpFixedHeaderLen {
		return nil, fmt.Errorf("helpfile: %s ptrBase inesperado: 0x%X", name, f.PtrBase)
	}

	// A tabela termina onde o primeiro ponteiro aponta (o texto começa
	// imediatamente após o último u32): N = (ptr[0] - ptrTableOff) / 4.
	firstPtr, err := readUint32(data, int(f.PtrTableOff))
	if err != nil {
		return nil, err
	}
	if firstPtr < f.PtrTableOff || firstPtr > f.TextEnd {
		return nil, fmt.Errorf("helpfile: %s primeiro ponteiro 0x%X inválido (ptrTableOff=0x%X textEnd=0x%X)",
			name, firstPtr, f.PtrTableOff, f.TextEnd)
	}
	if (firstPtr-f.PtrTableOff)%4 != 0 {
		return nil, fmt.Errorf("helpfile: %s tabela de ponteiros não alinhada (ptr[0]=0x%X ptrTableOff=0x%X)",
			name, firstPtr, f.PtrTableOff)
	}
	count := int((firstPtr - f.PtrTableOff) / 4)
	if count == 0 {
		return nil, fmt.Errorf("helpfile: %s sem ponteiros (ptr[0]=0x%X)", name, firstPtr)
	}

	f.Pointers = make([]uint32, count)
	for i := 0; i < count; i++ {
		f.Pointers[i] = binary.LittleEndian.Uint32(data[int(f.PtrTableOff)+i*4:])
	}
	for i, p := range f.Pointers {
		if p < firstPtr || p >= f.TextEnd {
			return nil, fmt.Errorf("helpfile: %s ponteiro %d (0x%X) fora da região de texto [0x%X,0x%X)",
				name, i, p, firstPtr, f.TextEnd)
		}
		if i > 0 && p < f.Pointers[i-1] {
			return nil, fmt.Errorf("helpfile: %s ponteiros não crescentes em %d (0x%X < 0x%X)",
				name, i, p, f.Pointers[i-1])
		}
	}

	if f.PtrTableOff > helpFixedHeaderLen {
		f.SubHeader = cloneBytes(data[helpFixedHeaderLen:int(f.PtrTableOff)])
	}

	f.Segments = make([]*HelpSegment, count)
	for i := 0; i < count; i++ {
		from := int(f.Pointers[i])
		to := int(f.TextEnd)
		if i < count-1 {
			to = int(f.Pointers[i+1])
		}
		if from >= to {
			return nil, fmt.Errorf("helpfile: %s segmento %d vazio (0x%X..0x%X)", name, i, from, to)
		}
		slice := data[from:to]

		term := bytes.IndexByte(slice, 0x00)
		if term < 0 {
			// Padrão do formato: todo texto tem o terminador. Falha aqui é
			// dado inesperado — loga e trata o segmento inteiro como texto
			// (round-trip preservado via RawBytes).
			common.LogError("helpfile: %s segmento %d sem terminador 0x00 (0x%X..0x%X)", name, i, from, to)
			term = len(slice) - 1
		}

		f.Segments[i] = &HelpSegment{
			Index:     i,
			RawBytes:  cloneBytes(slice[:term+1]),
			DataBytes: cloneBytes(slice[term+1:]),
			Text:      converter.BytesToString(slice[:term], localization, version),
		}
	}

	if f.FooterPtr != 0 {
		if f.FooterPtr < f.TextEnd || int(f.FooterPtr) > len(data) {
			return nil, fmt.Errorf("helpfile: %s footerPtr 0x%X inválido (textEnd=0x%X len=0x%X)",
				name, f.FooterPtr, f.TextEnd, len(data))
		}
		f.PagesBlock = cloneBytes(data[int(f.TextEnd):int(f.FooterPtr)])
		f.Footer = cloneBytes(data[int(f.FooterPtr):])

		if f.PageCount > 0 {
			recLen := int(f.PageCount) * 8
			if recLen > len(f.PagesBlock) {
				return nil, fmt.Errorf("helpfile: %s pageCount %d maior que o bloco de páginas (0x%X bytes)",
					name, f.PageCount, len(f.PagesBlock))
			}
			for r := 0; r < int(f.PageCount); r++ {
				sep := binary.LittleEndian.Uint16(f.PagesBlock[r*8+6:])
				if sep != HelpRecordSeparator {
					return nil, fmt.Errorf("helpfile: %s registro de página %d sem separador 0xFFFF", name, r)
				}
			}
		}
	} else if f.TextEnd != uint32(len(data)) {
		return nil, fmt.Errorf("helpfile: %s textEnd 0x%X não fecha o arquivo (len=0x%X) e footerPtr é 0",
			name, f.TextEnd, len(data))
	}

	return f, nil
}

// ReadHelpFile lê o .sps2 do painel de ajuda para versão/localização (resolução
// mods-first via NewFileAccessor) e parseia.
func ReadHelpFile(version common.GameVersion, localization, name string) (*HelpBinaryFile, error) {
	return ReadHelpFileFrom(version, localization, name, common.SourcePreferred)
}

// ReadHelpFileFrom é ReadHelpFile lendo da árvore indicada. SourceData
// entrega o painel pristine de data/ (coluna Original).
func ReadHelpFileFrom(version common.GameVersion, localization, name string, src common.FileSource) (*HelpBinaryFile, error) {
	path := HelpPathForVersion(version, localization, name)
	accessor, err := common.NewFileAccessorFrom(path, src)
	if err != nil {
		return nil, fmt.Errorf("helpfile: falha ao resolver %s: %w", path, err)
	}
	if !accessor.Exists {
		return nil, fmt.Errorf("helpfile: arquivo não encontrado (%s): %s", src, accessor.ResolvedPath)
	}
	data, err := accessor.ReadBytes()
	if err != nil {
		return nil, fmt.Errorf("helpfile: falha ao ler %s: %w", accessor.ResolvedPath, err)
	}
	f, err := ReadHelpBinary(data, name, localization, version)
	if err != nil {
		return nil, err
	}
	f.SourcePath = accessor.ResolvedPath
	return f, nil
}

// ReadHelpPanelFrom monta o painel completo (todas as localizações com
// arquivo) lendo da árvore indicada, SEM registrar no store. É o caminho do
// ORIGINAL: ReadHelpPanelFrom(version, name, common.SourceData).
func ReadHelpPanelFrom(version common.GameVersion, name string, src common.FileSource) *HelpKeyedStringFile {
	panel := NewHelpKeyedStringFile(name)
	for _, loc := range sortedSupportedLocalizations() {
		f, err := ReadHelpFileFrom(version, loc, name, src)
		if err != nil {
			continue
		}
		panel.Files[loc] = f
	}
	if len(panel.Files) == 0 {
		return nil
	}
	return panel
}
