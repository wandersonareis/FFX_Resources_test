package components

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type MacroString struct {
	Charset          string
	RegularOffset    int
	SimplifiedOffset int
	RegularBytes     []byte
	SimplifiedBytes  []byte
}

func NewMacroString(charset string, regularOffset, simplifiedOffset int, bytes []byte) *MacroString {
	regularBytes := GetStringBytesAtLookupOffset(bytes, regularOffset)

	var simplifiedBytes []byte
	if regularOffset == simplifiedOffset {
		simplifiedBytes = regularBytes
	} else {
		simplifiedBytes = GetStringBytesAtLookupOffset(bytes, simplifiedOffset)
	}

	return &MacroString{
		Charset:          charset,
		RegularOffset:    regularOffset,
		SimplifiedOffset: simplifiedOffset,
		RegularBytes:     regularBytes,
		SimplifiedBytes:  simplifiedBytes,
	}
}

func (m *MacroString) SetCharset(newCharset string) {
	if newCharset != "" && newCharset != m.Charset {
		m.Charset = newCharset
	}
}

func FromStringDataDev(data []byte, charset string) []*MacroString {
	if len(data) == 0 {
		return nil
	}

	// Cria um reader sobre o slice de bytes:
	r := bytes.NewReader(data)

	// Lê o primeiro uint16 (2 bytes) em little-endian. Esse valor indica
	// o offset do primeiro registro de string; ao dividir por 4 (tamanho de cada
	// par de offsets), obtemos a quantidade de strings (count).
	var first uint16
	if err := binary.Read(r, binary.LittleEndian, &first); err != nil {
		// Se der erro aqui, não há dados suficientes ou formato incorreto:
		fmt.Fprintf(os.Stderr, "Erro ao ler o primeiro offset: %v\n", err)
		return nil
	}
	count := int(first) / 4

	// Prealoca o slice de saída com capacidade exata:
	strings := make([]*MacroString, 0, count)

	// Se algo der panic dentro do loop, capturamos e exibimos no stderr:
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Fprintf(os.Stderr, "Exception during string data reading. (%v)\n", rec)
		}
	}()

	// Como já avançamos 2 bytes ao ler 'first', precisamos voltar ao início
	// para ler os pares (regularOffset, simplifiedOffset) de cada string:
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		fmt.Fprintf(os.Stderr, "Falha ao reposicionar o reader: %v\n", err)
		return nil
	}

	// Agora, a cada iteração, lemos dois uint16 (4 bytes) em little-endian:
	for i := 0; i < count; i++ {
		var regularOffset uint16
		var simplifiedOffset uint16

		// Lê os 2 bytes do offset regular:
		if err := binary.Read(r, binary.LittleEndian, &regularOffset); err != nil {
			panic(fmt.Sprintf("Erro lendo regularOffset no índice %d: %v", i, err))
		}

		// Lê os 2 bytes do offset simplificado:
		if err := binary.Read(r, binary.LittleEndian, &simplifiedOffset); err != nil {
			panic(fmt.Sprintf("Erro lendo simplifiedOffset no índice %d: %v", i, err))
		}

		// Converte para int e instancia o MacroString:
		strings = append(strings, NewMacroString(
			charset,
			int(regularOffset),
			int(simplifiedOffset),
			data,
		))
	}

	return strings
}

func FromStringData(bytes []byte, charset string) []*MacroString {
	if len(bytes) == 0 {
		return nil
	}

	first := int(bytes[0]) + int(bytes[1])<<8
	count := first / 4
	strings := make([]*MacroString, 0, count)

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Exception during string data reading. (%v)\n", r)
		}
	}()

	for i := 0; i < count; i++ {
		headerOffset := i * 4
		regularOffset := int(bytes[headerOffset]) + int(bytes[headerOffset+1])<<8
		simplifiedOffset := int(bytes[headerOffset+2]) + int(bytes[headerOffset+3])<<8
		strings = append(strings, NewMacroString(charset, regularOffset, simplifiedOffset, bytes))
	}

	return strings
}

func (m *MacroString) IsEmpty() bool {
	return m.GetRegularString() == "" && m.GetSimplifiedString() == ""
}

func (m *MacroString) GetRegularString() string {
	return BytesToString(m.RegularBytes, m.Charset)
}

func (m *MacroString) GetSimplifiedString() string {
	return BytesToString(m.SimplifiedBytes, m.Charset)
}

func (m *MacroString) HasDistinctSimplified() bool {
	if len(m.RegularBytes) != len(m.SimplifiedBytes) {
		return true
	}
	for i := range m.RegularBytes {
		if m.RegularBytes[i] != m.SimplifiedBytes[i] {
			return true
		}
	}
	return false
}

func (m *MacroString) GetString() string {
	if m.HasDistinctSimplified() {
		return fmt.Sprintf("%s (Simplified: %s)", m.GetRegularString(), m.GetSimplifiedString())
	}
	return m.GetRegularString()
}

func (m *MacroString) String() string {
	return m.GetString()
}
