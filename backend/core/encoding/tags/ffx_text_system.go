package tags

import "fmt"

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
	return []string{
		s.generateSystemCommand(),
		s.generateSystemWindowCode(),
		s.generateSystemAreaCode(),
	}
}

func (s *FFXTextTagSystem) generateSystemCommand() string {
	return fmt.Sprintf("\\x%02X\\c%02X={x%02X:\\h%02X}", s.systemByte, s.systemTag, s.systemByte, s.systemTag)
}

func (s *FFXTextTagSystem) generateSystemWindowCode() string {
	return fmt.Sprintf("\\x%02X\\x%02X={%s}%s", s.systemByte, 0x30, "WINDOW", LineBreakString())
}

func (s *FFXTextTagSystem) generateSystemAreaCode() string {
	return fmt.Sprintf("\\x%02X\\x%02X={%s}", s.systemByte, 0x3F, "AREA")
}
