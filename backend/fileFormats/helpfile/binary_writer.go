package helpfile

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
)

// RebuildHelpBinary reconstrói o binário .sps2 a partir do arquivo carregado.
//
// Regras do formato:
//   - cada segmento grava o texto (via converter.StringToStoredBytes, que
//     inclui o terminador 0x00) + DataBytes; texto inalterado usa RawBytes
//     verbatim (round-trip byte-exato);
//   - ponteiros recalculados pela posição acumulada: ptr[i] = textStart +
//     soma dos segmentos anteriores — a regra "len do texto após
//     FillByteList + bytes de dados = ponteiro recalculado";
//   - textEnd (@0x08) e footerPtr (@0x14) recalculados pelo delta do texto;
//     pageCount (@0x04), ptrBase (@0x0C) e ptrTableOff (@0x10) são estáticos;
//   - bloco de páginas preservado com os offsets dos registros deslocados
//     pelo mesmo delta; footer preservado verbatim.
func RebuildHelpBinary(f *HelpBinaryFile, charset string, version common.GameVersion) ([]byte, error) {
	if f == nil {
		return nil, fmt.Errorf("helpfile: arquivo nulo")
	}
	if len(f.Segments) != len(f.Pointers) {
		return nil, fmt.Errorf("helpfile: %s segmentos (%d) != ponteiros (%d)",
			f.Name, len(f.Segments), len(f.Pointers))
	}

	textStart := int(f.PtrTableOff) + 4*len(f.Pointers)
	if textStart <= helpFixedHeaderLen {
		return nil, fmt.Errorf("helpfile: %s início de texto inválido (0x%X)", f.Name, textStart)
	}

	var buf bytes.Buffer
	buf.Grow(4 * len(f.Segments))
	newPointers := make([]uint32, len(f.Segments))

	for i, seg := range f.Segments {
		newPointers[i] = uint32(textStart + buf.Len())
		if !seg.dirty && len(seg.RawBytes) > 0 {
			buf.Write(seg.RawBytes)
		} else {
			encoded, err := converter.StringToStoredBytes(seg.Text, charset, version)
			if err != nil {
				return nil, fmt.Errorf("helpfile: %s falha ao encodar segmento %d: %w", f.Name, i, err)
			}
			buf.Write(encoded)
			// RawBytes passa a valer a reencoding atual: depois que o Save
			// limpar o dirty, o próximo rebuild usa estes bytes (e não os
			// originais) — senão a edição regravação seguinte reverteria.
			seg.RawBytes = encoded
		}
		buf.Write(seg.DataBytes)
	}

	newTextEnd := uint32(textStart + buf.Len())
	delta := int(newTextEnd) - int(f.TextEnd)

	footerPtr := f.FooterPtr
	pagesBlock := f.PagesBlock
	if f.PageCount > 0 || len(f.PagesBlock) > 0 {
		footerPtr = uint32(int(f.FooterPtr) + delta)
		pagesBlock = shiftPageRecordOffsets(f.PagesBlock, int(f.PageCount), delta)
	}

	out := bytes.NewBuffer(make([]byte, 0, textStart+buf.Len()+len(pagesBlock)+len(f.Footer)))
	writeHeader(out, f.Magic, f.PageCount, newTextEnd, f.PtrBase, f.PtrTableOff, footerPtr, f.TimeStamp)
	out.Write(f.SubHeader)
	for _, p := range newPointers {
		binary.Write(out, binary.LittleEndian, p)
	}
	out.Write(buf.Bytes())
	out.Write(pagesBlock)
	out.Write(f.Footer)

	return out.Bytes(), nil
}

func writeHeader(out *bytes.Buffer, magic, pageCount, textEnd, ptrBase, ptrTableOff, footerPtr uint32, timeStamp [12]byte) {
	binary.Write(out, binary.LittleEndian, magic)
	binary.Write(out, binary.LittleEndian, pageCount)
	binary.Write(out, binary.LittleEndian, textEnd)
	binary.Write(out, binary.LittleEndian, ptrBase)
	binary.Write(out, binary.LittleEndian, ptrTableOff)
	binary.Write(out, binary.LittleEndian, footerPtr)
	out.Write(timeStamp[:])
}

// shiftPageRecordOffsets desloca os offsets (u32 no início de cada registro
// de 8 bytes) pelo delta do texto: o bloco de páginas anda inteiro quando o
// texto cresce ou encolhe, e os offsets são absolutos no arquivo.
func shiftPageRecordOffsets(pages []byte, pageCount, delta int) []byte {
	if delta == 0 || pageCount <= 0 || len(pages) == 0 {
		return pages
	}
	out := make([]byte, len(pages))
	copy(out, pages)
	for r := 0; r < pageCount; r++ {
		off := binary.LittleEndian.Uint32(out[r*8:])
		binary.LittleEndian.PutUint32(out[r*8:], off+uint32(delta))
	}
	return out
}

// Rebuild aplica as alterações de texto e devolve o binário pronto, usando o
// charset da localização do arquivo.
func (f *HelpBinaryFile) Rebuild() ([]byte, error) {
	return RebuildHelpBinary(f, f.Charset(), f.Version)
}

// Save reconstrói o binário e grava no padrão dos demais formatos: a árvore
// mods/ (mods/<localizationRoot>/help/<dir>/<name>.sps2). O binário original
// do gamefiles permanece intacto — a leitura é mods-first, então o mod
// passa a ser a fonte nas próximas cargas.
func (f *HelpBinaryFile) Save() error {
	if f.Name == "" {
		return fmt.Errorf("helpfile: arquivo sem nome (use ReadHelpFile)")
	}
	target := HelpModsPathForVersion(f.Version, f.Charset(), f.Name)

	data, err := f.Rebuild()
	if err != nil {
		return err
	}

	if err := common.WriteBytesToFile(target, data); err != nil {
		return fmt.Errorf("helpfile: falha ao gravar %s: %w", target, err)
	}
	// Gravação bem-sucedida: os segmentos voltam a "limpos" (RawBytes já
	// vale a reencoding atual, ver RebuildHelpBinary). Assim um Save
	// seguinte sem novas alterações é no-op de verdade.
	for _, seg := range f.Segments {
		seg.dirty = false
	}

	common.LogInfo("helpfile: %s gravado (%d bytes, textEnd=0x%X)", target, len(data),
		binary.LittleEndian.Uint32(data[0x08:]))
	return nil
}
