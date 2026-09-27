package helpfile

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	ffxencoding "ffxresources/backend/core/encoding"
)

// Fixtures dos painéis de ajuda em testData (mesma árvore new_uspc do jogo).
var (
	testNowHelpPath     = filepath.Join("..", "..", "..", "testData", "FFX", "binary", "ffx_ps2", "ffx", "master", "new_uspc", "help", "now_help", "now_help.sps2")
	testNowHelpPagePath = filepath.Join("..", "..", "..", "testData", "FFX", "binary", "ffx_ps2", "ffx", "master", "new_uspc", "help", "now_help", "now_help_page.sps2")
	testMonBokuPath     = filepath.Join("..", "..", "..", "testData", "FFX", "binary", "ffx_ps2", "ffx", "master", "new_uspc", "help", "mon_boku", "mon_boku.sps2")
	testSMonitorPath    = filepath.Join("..", "..", "..", "testData", "FFX", "binary", "ffx_ps2", "ffx", "master", "new_uspc", "help", "s_monitor", "s_monitor.sps2")
	testDvdcopyPath     = filepath.Join("..", "..", "..", "testData", "FFX", "binary", "ffx_ps2", "ffx", "master", "new_uspc", "help", "dvdcopy", "dvdcopy.sps2")
	testDvdcopyPagePath = filepath.Join("..", "..", "..", "testData", "FFX", "binary", "ffx_ps2", "ffx", "master", "new_uspc", "help", "dvdcopy", "dvdcopy_page.sps2")
)

func loadFixture(t *testing.T, path, name string) ([]byte, *HelpBinaryFile) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("falha ao ler fixture %s: %v", path, err)
	}
	if err := ffxencoding.PrepareVersionCharsets(common.GameVersionFFX); err != nil {
		t.Fatalf("charset maps não carregados: %v", err)
	}
	f, err := ReadHelpBinary(data, name, common.DefaultLocalization, common.GameVersionFFX)
	if err != nil {
		t.Fatalf("falha ao parsear %s: %v", name, err)
	}
	return data, f
}

func TestReadNowHelp(t *testing.T) {
	data, f := loadFixture(t, testNowHelpPath, NowHelpName)

	if len(f.Segments) != 1208 {
		t.Fatalf("segmentos: got %d want 1208", len(f.Segments))
	}
	if f.TextEnd != uint32(len(data)) {
		t.Fatalf("textEnd: got 0x%X want 0x%X (arquivo sem bloco de páginas)", f.TextEnd, len(data))
	}
	if f.PageCount != 0 || f.PtrBase != 0 {
		t.Fatalf("header simples: pageCount=%d ptrBase=0x%X (esperado 0/0)", f.PageCount, f.PtrBase)
	}
	if f.PtrTableOff != 0x24 {
		t.Fatalf("ptrTableOff: got 0x%X want 0x24", f.PtrTableOff)
	}
	if f.FooterPtr != 0 {
		t.Fatalf("footerPtr: got 0x%X want 0", f.FooterPtr)
	}
	if f.Pointers[0] != 0x1304 || f.Pointers[1] != 0x130B || f.Pointers[8] != 0x1488 {
		t.Fatalf("ponteiros iniciais inesperados: %v", f.Pointers[:9])
	}

	// Todo segmento tem o terminador: RawBytes termina em 0x00.
	for i, seg := range f.Segments {
		if len(seg.RawBytes) == 0 || seg.RawBytes[len(seg.RawBytes)-1] != 0x00 {
			t.Fatalf("segmento %d sem terminador 0x00 em RawBytes", i)
		}
		if seg.Index != i {
			t.Fatalf("segmento %d com Index %d", i, seg.Index)
		}
	}
}

func TestReadNowHelpPage(t *testing.T) {
	data, f := loadFixture(t, testNowHelpPagePath, NowHelpPageName)

	if len(f.Segments) != 1208 {
		t.Fatalf("segmentos: got %d want 1208", len(f.Segments))
	}
	if f.PageCount != 155 {
		t.Fatalf("pageCount: got %d want 155", f.PageCount)
	}
	if f.PtrBase != 0x24 || f.PtrTableOff != 0x114 {
		t.Fatalf("header _page: ptrBase=0x%X ptrTableOff=0x%X (esperado 0x24/0x114)", f.PtrBase, f.PtrTableOff)
	}
	if f.TextEnd != 0xAE00 {
		t.Fatalf("textEnd: got 0x%X want 0xAE00", f.TextEnd)
	}
	if f.FooterPtr != 0xBD7F0 {
		t.Fatalf("footerPtr: got 0x%X want 0xBD7F0", f.FooterPtr)
	}
	if want := int(0xBD7F0 - 0xAE00); len(f.PagesBlock) != want {
		t.Fatalf("PagesBlock: got %d want %d", len(f.PagesBlock), want)
	}
	if want := len(data) - int(0xBD7F0); len(f.Footer) != want {
		t.Fatalf("Footer: got %d want %d", len(f.Footer), want)
	}

	// Os 155 registros de página têm o separador 0xFFFF e offsets dentro do
	// próprio bloco (entre textEnd e footerPtr).
	for r := 0; r < 155; r++ {
		off := int(bytesToU32(f.PagesBlock[r*8:]))
		if off < int(f.TextEnd) || off >= int(f.FooterPtr) {
			t.Fatalf("registro %d com offset 0x%X fora do bloco de páginas", r, off)
		}
	}
}

func TestSegmentTextsDecoded(t *testing.T) {
	_, f := loadFixture(t, testNowHelpPath, NowHelpName)

	// seg 3: "Directional Button SE" seguido do código de ícone 0B F0 na
	// porção pré-terminador (decode {ICON:F0:?} re-encoda 0B F0).
	if got := f.Segments[3].Text; got != "Directional Button SE{ICON:F0:?}" {
		t.Fatalf("seg 3: got %q want %q", got, "Directional Button SE{ICON:F0:?}")
	}
	if len(f.Segments[3].DataBytes) == 0 {
		t.Fatal("seg 3 devia ter bloco de dados após o terminador")
	}
	// seg 2: " {TEXT_NEWLINE}" (espaço + quebra de linha 0x03, texto mínimo
	// terminado em 0x00).
	if got := f.Segments[2].Text; got != " {TEXT_NEWLINE}" {
		t.Fatalf("seg 2: got %q want %q", got, " {TEXT_NEWLINE}")
	}
	// seg 4: template "Battles fought: ... Times KO'd: ... Enemies defeated:"
	// (frase de encaixe com espaços de padding).
	if got := f.Segments[4].Text; !strings.Contains(got, "Battles fought") || !strings.Contains(got, "Enemies defeated") {
		t.Fatalf("seg 4: texto de template inesperado: %q", got)
	}
	if got := f.Segments[4].DataBytes; len(got) != 0 {
		t.Fatalf("seg 4 devia ser texto puro, DataBytes=%d", len(got))
	}
	// seg 0: "Dummy" + quebra de linha codificada.
	if got := f.Segments[0].Text; got == "" {
		t.Fatal("seg 0 não deveria ser vazio")
	}
}

func TestRoundTripNowHelp(t *testing.T) {
	original, f := loadFixture(t, testNowHelpPath, NowHelpName)

	rebuilt, err := RebuildHelpBinary(f, f.Charset(), f.Version)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if !bytes.Equal(rebuilt, original) {
		t.Fatal("round-trip do now_help.sps2 divergiu do original byte a byte")
	}
}

func TestRoundTripNowHelpPage(t *testing.T) {
	original, f := loadFixture(t, testNowHelpPagePath, NowHelpPageName)

	rebuilt, err := RebuildHelpBinary(f, f.Charset(), f.Version)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if !bytes.Equal(rebuilt, original) {
		t.Fatal("round-trip do now_help_page.sps2 divergiu do original byte a byte")
	}
}

func TestRoundTripMonBoku(t *testing.T) {
	original, f := loadFixture(t, testMonBokuPath, MonBokuName)

	if len(f.Segments) != 1578 {
		t.Fatalf("segmentos: got %d want 1578", len(f.Segments))
	}
	// mon_boku é o caso pesado de preservação: 71% dos segmentos têm bloco
	// de dados (blocos de status dos monstros).
	dataBlocks := 0
	for _, seg := range f.Segments {
		if len(seg.DataBytes) > 0 {
			dataBlocks++
		}
	}
	if dataBlocks != 1128 {
		t.Fatalf("blocos de dados: got %d want 1128", dataBlocks)
	}

	rebuilt, err := RebuildHelpBinary(f, f.Charset(), f.Version)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if !bytes.Equal(rebuilt, original) {
		t.Fatal("round-trip do mon_boku.sps2 divergiu do original byte a byte")
	}
}

func TestRoundTripSMonitor(t *testing.T) {
	original, f := loadFixture(t, testSMonitorPath, SMonitorName)

	if len(f.Segments) != 197 {
		t.Fatalf("segmentos: got %d want 197", len(f.Segments))
	}
	rebuilt, err := RebuildHelpBinary(f, f.Charset(), f.Version)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if !bytes.Equal(rebuilt, original) {
		t.Fatal("round-trip do s_monitor.sps2 divergiu do original byte a byte")
	}
}

func TestRoundTripDvdcopy(t *testing.T) {
	original, f := loadFixture(t, testDvdcopyPath, DvdcopyName)

	if len(f.Segments) != 70 {
		t.Fatalf("segmentos: got %d want 70", len(f.Segments))
	}
	rebuilt, err := RebuildHelpBinary(f, f.Charset(), f.Version)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if !bytes.Equal(rebuilt, original) {
		t.Fatal("round-trip do dvdcopy.sps2 divergiu do original byte a byte")
	}
}

func TestRoundTripDvdcopyPage(t *testing.T) {
	original, f := loadFixture(t, testDvdcopyPagePath, DvdcopyPageName)

	if len(f.Segments) == 0 {
		t.Fatal("dvdcopy_page sem segmentos")
	}
	if f.PageCount != 6 || f.PtrTableOff != 0x84 {
		t.Fatalf("header _page inesperado: pageCount=%d ptrTableOff=0x%X", f.PageCount, f.PtrTableOff)
	}
	rebuilt, err := RebuildHelpBinary(f, f.Charset(), f.Version)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	if !bytes.Equal(rebuilt, original) {
		t.Fatal("round-trip do dvdcopy_page.sps2 divergiu do original byte a byte")
	}
}

func TestRebuildWithModifiedText(t *testing.T) {
	_, f := loadFixture(t, testNowHelpPath, NowHelpName)

	// Modifica um texto puro (seg 3) encurtando-o: ponteiros seguintes devem
	// recuar e textEnd recalcular. Captura o tamanho original de RawBytes
	// antes do rebuild (o rebuild reatribui RawBytes do segmento sujo).
	origRawLen := len(f.Segments[3].RawBytes)
	f.Segments[3].SetText("Button")
	rebuilt, err := RebuildHelpBinary(f, f.Charset(), f.Version)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	reread, err := ReadHelpBinary(rebuilt, NowHelpName, common.DefaultLocalization, common.GameVersionFFX)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	if got := reread.Segments[3].Text; got != "Button" {
		t.Fatalf("seg 3 pós-import: got %q want %q", got, "Button")
	}
	if len(reread.Segments) != 1208 {
		t.Fatalf("segmentos pós-import: %d", len(reread.Segments))
	}

	// Seg 3 original: texto+icon (22 bytes) + 0B F0 (2) + 0x00 = 24 bytes;
	// novo: "Button" (6) + 0x00 = 7 → delta -17, calculado do próprio arquivo.
	newEncoded, err := converter.StringToStoredBytes("Button", f.Charset(), f.Version)
	if err != nil {
		t.Fatalf("encode do novo texto: %v", err)
	}
	deltaWant := len(newEncoded) - origRawLen
	delta := int(reread.Pointers[4]) - int(f.Pointers[4])
	if delta != deltaWant {
		t.Fatalf("delta do ponteiro seguinte: got %d want %d", delta, deltaWant)
	}
	// Ponteiros anteriores ao segmento editado permanecem.
	if reread.Pointers[3] != f.Pointers[3] {
		t.Fatalf("ponteiro anterior ao segmento editado mudou: 0x%X -> 0x%X", f.Pointers[3], reread.Pointers[3])
	}
	if reread.TextEnd != uint32(len(rebuilt)) {
		t.Fatalf("textEnd pós-import 0x%X != tamanho do arquivo 0x%X", reread.TextEnd, len(rebuilt))
	}
}

func TestRebuildPageWithModifiedText(t *testing.T) {
	original, f := loadFixture(t, testNowHelpPagePath, NowHelpPageName)

	// Encurta um texto puro: o bloco de páginas anda para trás e os offsets
	// dos 155 registros acompanham o delta.
	oldPtr := f.Pointers[4]
	f.Segments[4].SetText("OK")
	rebuilt, err := RebuildHelpBinary(f, f.Charset(), f.Version)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}

	reread, err := ReadHelpBinary(rebuilt, NowHelpPageName, common.DefaultLocalization, common.GameVersionFFX)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}

	newTextEnd := bytesToU32(rebuilt[0x08:])
	if newTextEnd >= f.TextEnd {
		t.Fatalf("textEnd deveria encolher: got 0x%X (original 0x%X)", newTextEnd, f.TextEnd)
	}
	delta := int(newTextEnd) - int(f.TextEnd)

	// Ponteiro do segmento editado fica no lugar; o seguinte recua.
	if reread.Pointers[4] != oldPtr {
		t.Fatalf("ponteiro do segmento editado mudou: 0x%X -> 0x%X", oldPtr, reread.Pointers[4])
	}
	if got := int(reread.Pointers[5]) - int(f.Pointers[5]); got != delta {
		t.Fatalf("delta do ponteiro seguinte: got %d want %d", got, delta)
	}
	// Bloco de páginas inteiro deslocado pelo mesmo delta.
	if got := int(bytesToU32(reread.PagesBlock[0:])) - int(bytesToU32(f.PagesBlock[0:])); got != delta {
		t.Fatalf("delta do registro 0 do bloco de páginas: got %d want %d", got, delta)
	}
	// FooterPtr recalculado pelo mesmo delta.
	if got := int(reread.FooterPtr) - int(f.FooterPtr); got != delta {
		t.Fatalf("delta do footerPtr: got %d want %d", got, delta)
	}
	// Bloco de páginas preservado: conteúdo idêntico exceto os offsets dos
	// 155 registros (deslocados pelo delta).
	wantPages := shiftPageRecordOffsets(f.PagesBlock, int(f.PageCount), delta)
	if !bytes.Equal(reread.PagesBlock, wantPages) {
		t.Fatal("bloco de páginas divergiu além do deslocamento de offsets")
	}
	if !bytes.Equal(reread.Footer, f.Footer) {
		t.Fatal("footer divergiu do original")
	}
	if !bytes.Equal(rebuilt[int(reread.FooterPtr):], original[int(f.FooterPtr):]) {
		t.Fatal("footer no binário reconstruído divergiu do original")
	}
	// Header estático intacto.
	if reread.PageCount != 155 || reread.PtrBase != 0x24 || reread.PtrTableOff != 0x114 {
		t.Fatalf("header estático alterado: pageCount=%d ptrBase=0x%X ptrTableOff=0x%X",
			reread.PageCount, reread.PtrBase, reread.PtrTableOff)
	}
}

func bytesToU32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}
