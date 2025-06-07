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

func FromStringData(data []byte, charset string) []*MacroString {
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

func (m *MacroString) IsEmpty() bool {
	return m.GetRegularString() == "" && m.GetSimplifiedString() == ""
}

func (m *MacroString) GetRegularString() string {
	return BytesToString(m.RegularBytes, m.Charset)
}

func (m *MacroString) GetRegularBytes() []byte {
	return m.RegularBytes
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

func RebuildMacroStrings(strings []*MacroString, charset string, optimize bool) {
	count := len(strings)
	contentOffset := count * 4 // Each macro string uses 4 bytes (2 for regular offset, 2 for simplified offset)
	offsetMap := make(map[string]int)
	var buf bytes.Buffer

	// Build the content section and calculate offsets
	for _, macroString := range strings {
		if macroString == nil {
			continue
		}

		// Handle regular string
		regularString := macroString.GetRegularString()
		if regularString == "" {
			macroString.RegularOffset = contentOffset
		} else if optimize {
			if offset, exists := offsetMap[regularString]; exists {
				macroString.RegularOffset = contentOffset + offset
			} else {
				macroString.RegularOffset = contentOffset + buf.Len()
				offsetMap[regularString] = buf.Len()
				regularBytes := StringToBytes(regularString, charset)
				buf.Write(regularBytes)
				buf.WriteByte(0x00) // null terminator
			}
		} else {
			macroString.RegularOffset = contentOffset + buf.Len()
			regularBytes := StringToBytes(regularString, charset)
			buf.Write(regularBytes)
			buf.WriteByte(0x00) // null terminator
		}

		// Handle simplified string
		simplifiedString := macroString.GetSimplifiedString()
		if simplifiedString == "" {
			macroString.SimplifiedOffset = contentOffset
		} else if optimize && regularString == simplifiedString {
			// If simplified is same as regular, use same offset
			macroString.SimplifiedOffset = macroString.RegularOffset
		} else if optimize {
			if offset, exists := offsetMap[simplifiedString]; exists {
				macroString.SimplifiedOffset = contentOffset + offset
			} else {
				macroString.SimplifiedOffset = contentOffset + buf.Len()
				offsetMap[simplifiedString] = buf.Len()
				simplifiedBytes := StringToBytes(simplifiedString, charset)
				buf.Write(simplifiedBytes)
				buf.WriteByte(0x00) // null terminator
			}
		} else {
			macroString.SimplifiedOffset = contentOffset + buf.Len()
			simplifiedBytes := StringToBytes(simplifiedString, charset)
			buf.Write(simplifiedBytes)
			buf.WriteByte(0x00) // null terminator
		}
	}
}
func GenerateMacroStringData(strings []*MacroString, charset string, optimize bool) []byte {
	count := len(strings)
	contentOffset := count * 4 // Each macro string uses 4 bytes (2 for regular offset, 2 for simplified offset)
	offsetMap := make(map[string]int)
	var buf bytes.Buffer

	// Build the content section and calculate offsets
	for _, macroString := range strings {
		if macroString == nil {
			continue
		}

		// Handle regular string
		regularString := macroString.GetRegularString()
		if regularString == "" {
			macroString.RegularOffset = contentOffset
		} else if optimize {
			if offset, exists := offsetMap[regularString]; exists {
				macroString.RegularOffset = contentOffset + offset
			} else {
				macroString.RegularOffset = contentOffset + buf.Len()
				offsetMap[regularString] = buf.Len()
				regularBytes := StringToBytes(regularString, charset)
				buf.Write(regularBytes)
				buf.WriteByte(0x00) // null terminator
			}
		} else {
			macroString.RegularOffset = contentOffset + buf.Len()
			regularBytes := StringToBytes(regularString, charset)
			buf.Write(regularBytes)
			buf.WriteByte(0x00) // null terminator
		}

		// Handle simplified string
		simplifiedString := macroString.GetSimplifiedString()
		if simplifiedString == "" {
			macroString.SimplifiedOffset = contentOffset
		} else if optimize && regularString == simplifiedString {
			// If simplified is same as regular, use same offset
			macroString.SimplifiedOffset = macroString.RegularOffset
		} else if optimize {
			if offset, exists := offsetMap[simplifiedString]; exists {
				macroString.SimplifiedOffset = contentOffset + offset
			} else {
				macroString.SimplifiedOffset = contentOffset + buf.Len()
				offsetMap[simplifiedString] = buf.Len()
				simplifiedBytes := StringToBytes(simplifiedString, charset)
				buf.Write(simplifiedBytes)
				buf.WriteByte(0x00) // null terminator
			}
		} else {
			macroString.SimplifiedOffset = contentOffset + buf.Len()
			simplifiedBytes := StringToBytes(simplifiedString, charset)
			buf.Write(simplifiedBytes)
			buf.WriteByte(0x00) // null terminator
		}
	}

	// Now build the final byte array with headers + content
	totalSize := count*4 + buf.Len()
	result := make([]byte, totalSize)

	// Write the header section (offsets)
	for i, macroString := range strings {
		offset := i * 4
		if macroString == nil {
			// Write zeros for nil entries
			result[offset] = 0
			result[offset+1] = 0
			result[offset+2] = 0
			result[offset+3] = 0
		} else {
			// Write regular offset (little-endian 16-bit)
			result[offset] = byte(macroString.RegularOffset & 0xFF)
			result[offset+1] = byte((macroString.RegularOffset >> 8) & 0xFF)

			// Write simplified offset (little-endian 16-bit)
			result[offset+2] = byte(macroString.SimplifiedOffset & 0xFF)
			result[offset+3] = byte((macroString.SimplifiedOffset >> 8) & 0xFF)
		}
	}

	// Copy the content section
	copy(result[count*4:], buf.Bytes())

	return result
}

func MacroStringsToBytes(strings []*MacroString, charset string, optimize bool) []byte {
	if len(strings) == 0 {
		return []byte{}
	}

	// Build the string data section
	stringData := GenerateMacroStringData(strings, charset, optimize)

	// Calculate the total size needed: 2 bytes for count + string data
	totalSize := 2 + len(stringData)
	result := make([]byte, totalSize)

	// Write the count as first 2 bytes (little-endian)
	count := len(strings) * 4 // Each entry uses 4 bytes in the header
	result[0] = byte(count & 0xFF)
	result[1] = byte((count >> 8) & 0xFF)

	// Copy the string data
	copy(result[2:], stringData)

	return result
}

func (m *MacroString) SetRegularString(str string, newCharset ...string) {
	if len(newCharset) > 0 && newCharset[0] != "" {
		m.SetCharset(newCharset[0])
	}

	keepSimplifiedSynced := !m.HasDistinctSimplified()
	m.RegularBytes = StringToBytes(str, m.Charset)

	if keepSimplifiedSynced {
		m.SimplifiedBytes = m.RegularBytes
	}
}

func (m *MacroString) SetSimplifiedString(str string, newCharset ...string) {
	if len(newCharset) > 0 && newCharset[0] != "" {
		m.SetCharset(newCharset[0])
	}

	m.SimplifiedBytes = StringToBytes(str, m.Charset)
}

func (m *MacroString) SetCharset(newCharset string) {
	if newCharset != "" && newCharset != m.Charset {
		m.Charset = newCharset
	}
}
