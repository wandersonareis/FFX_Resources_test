package chunkmap

import (
	"encoding/binary"
	"fmt"

	"ffxresources/backend/common"
)

// FileHeader é o cabeçalho dos binários de objetos textuais, nas duas
// versões que objectsfile lê:
//
// V1 — FFX, 20 bytes (0x14):
//
//	sigA u32 @0x00, sigB u32 @0x04, min u16 @0x08, max u16 @0x0A,
//	individual u16 @0x0C, total u16 @0x0E, padding u32 @0x10
//
// V2 — FFX-2 / LastMission, 32 bytes (0x20):
//
//	sigA u32 @0x00, sigB u32 @0x04, unknown u32 @0x08, min u32 @0x0C,
//	max u32 @0x10, individual u32 @0x14, total u32 @0x18,
//	headerSize u32 @0x1C
//
// Os chunks começam em HeaderLength e ocupam TotalLength bytes; a string
// table é todo o resto do arquivo (offset relativo a Chunk.Offset).
type FileHeader struct {
	HeaderLength     int
	MinIndex         int
	MaxIndex         int
	IndividualLength int
	TotalLength      int
}

// Count devolve a quantidade de objetos declarada no cabeçalho: max - min + 1.
// Nunca derivar de len/ind — o header é a única fonte confiável.
func (h FileHeader) Count() int { return h.MaxIndex - h.MinIndex + 1 }

// ChunkCount devolve quantos chunks COMPLETOS cabem em TotalLength — o leitor
// real quebra o loop quando o chunk estoura a faixa (mesma regra de
// ReadDataListWithIlist/V2).
func (h FileHeader) ChunkCount() int {
	n := h.Count()
	if h.IndividualLength > 0 {
		if max := h.TotalLength / h.IndividualLength; n > max {
			return max
		}
	}
	return n
}

// Chunks devolve a faixa de bytes dos chunks dentro de data.
func (h FileHeader) Chunks(data []byte) []byte {
	return data[h.HeaderLength : h.HeaderLength+h.TotalLength]
}

// StringTable devolve o resto do arquivo depois dos chunks.
func (h FileHeader) StringTable(data []byte) []byte {
	return data[h.HeaderLength+h.TotalLength:]
}

// Chunk devolve o i-ésimo chunk dentro de Chunks(data).
func (h FileHeader) Chunk(chunks []byte, i int) []byte {
	from := i * h.IndividualLength
	return chunks[from : from+h.IndividualLength]
}

// ReadFileHeader lê e valida o cabeçalho de data. A regra V1/V2 é a mesma de
// objectsfile.PopulateDataObjectLocalizationsWithIlist: só FFX usa V1.
func ReadFileHeader(data []byte, version common.GameVersion) (FileHeader, error) {
	var h FileHeader
	if version == common.GameVersionFFX {
		if len(data) < 20 {
			return h, fmt.Errorf("%s: %d bytes (< 20)", version, len(data))
		}
		h.HeaderLength = 20
		h.MinIndex = int(binary.LittleEndian.Uint16(data[8:10]))
		h.MaxIndex = int(binary.LittleEndian.Uint16(data[10:12]))
		h.IndividualLength = int(binary.LittleEndian.Uint16(data[12:14]))
		h.TotalLength = int(binary.LittleEndian.Uint16(data[14:16]))
	} else {
		if len(data) < 32 {
			return h, fmt.Errorf("%s: %d bytes (< 32)", version, len(data))
		}
		h.HeaderLength = 32
		h.MinIndex = int(binary.LittleEndian.Uint32(data[12:16]))
		h.MaxIndex = int(binary.LittleEndian.Uint32(data[16:20]))
		h.IndividualLength = int(binary.LittleEndian.Uint32(data[20:24]))
		h.TotalLength = int(binary.LittleEndian.Uint32(data[24:28]))
	}
	if h.IndividualLength <= 0 || h.Count() <= 0 {
		return h, fmt.Errorf("%s: cabecalho invalido: min=%d max=%d individual=%d",
			version, h.MinIndex, h.MaxIndex, h.IndividualLength)
	}
	if h.HeaderLength+h.TotalLength > len(data) {
		return h, fmt.Errorf("%s: header+chunks (%d) maior que o arquivo (%d)",
			version, h.HeaderLength+h.TotalLength, len(data))
	}
	return h, nil
}

// Split lê o cabeçalho e separa data em chunks e string table.
func Split(data []byte, version common.GameVersion) (FileHeader, []byte, []byte, error) {
	h, err := ReadFileHeader(data, version)
	if err != nil {
		return h, nil, nil, err
	}
	return h, h.Chunks(data), h.StringTable(data), nil
}
