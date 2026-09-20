package macrodic

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	"fmt"
)

// MacroString is a single macro string entry: the Regular/Simplified offsets
// (uint16) into its file bytes plus the resolved text bytes. When both offsets
// are equal, both fields point to the same text (no distinct simplified text).
type MacroString struct {
	Charset          string
	Version          common.GameVersion
	RegularOffset    int
	SimplifiedOffset int
	RegularBytes     []byte
	SimplifiedBytes  []byte
}

func (m *MacroString) IsEmpty() bool {
	return m.GetRegularString() == "" && m.GetSimplifiedString() == ""
}

func (m *MacroString) GetRegularString() string {
	return converter.BytesToString(m.RegularBytes, m.Charset, m.Version)
}

func (m *MacroString) GetSimplifiedString() string {
	return converter.BytesToString(m.SimplifiedBytes, m.Charset, m.Version)
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
