package tags

import (
	"fmt"
	"slices"
)

type FFXTextTagSystem struct {
	systemByte byte
	systemTag  byte
}

func NewTextTagSystem() *FFXTextTagSystem {
	return &FFXTextTagSystem{
		systemByte: 0x09,
		systemTag:  0x01,
	}
}

func (s *FFXTextTagSystem) FFXTextSystemCodePage() []string {
	return slices.Concat(
		s.generateSystemCommand(),
		s.generateSystemCodePage(),
	)
}

func (s *FFXTextTagSystem) generateSystemCommand() []string {
	systemCommand := fmt.Sprintf("\\x%02X\\c%02X={x%02X:\\h%02X}", s.systemByte, s.systemTag, s.systemByte, s.systemTag)

	return []string{systemCommand}
}

func (s *FFXTextTagSystem) generateSystemCodePage() []string {
	systemWindow := fmt.Sprintf("\\x%02X\\x%02X={%s}%s", s.systemByte, 0x30, "WINDOW", LineBreakString())

	systemArea := fmt.Sprintf("\\x%02X\\x%02X={%s}", s.systemByte, 0x3F, "AREA")

	return []string{
		systemWindow,
		systemArea,
	}
}
