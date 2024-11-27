package tags

import "fmt"

type FFXTextTagSystem struct {
	systemByte byte
	systemTag  byte
}

func NewTextTagSystem() *FFXTextTagSystem {
	systemByte := byte(0x09)
	systemTag := byte(0x01)

	return &FFXTextTagSystem{
		systemByte: systemByte,
		systemTag:  systemTag,
	}
}

func (s *FFXTextTagSystem) FFXTextSystemCodePage() []string {
	systems := make([]string, 0, 3)

	s.generateSystemCommand(&systems)

	s.generateSystemCodePage(&systems)

	return systems
}

func (s *FFXTextTagSystem) generateSystemCommand(list *[]string) {
	systemCommand := fmt.Sprintf("\\x%02X\\c%02X={x%02X:\\h%02X}", s.systemByte, s.systemTag, s.systemByte, s.systemTag)

	*list = append(*list, systemCommand)
}

func (s *FFXTextTagSystem) generateSystemCodePage(list *[]string) {
	systemWindow := fmt.Sprintf("\\x%02X\\x%02X={%s}%s", s.systemByte, 0x30, "WINDOW", LineBreakString())

	systemArea := fmt.Sprintf("\\x%02X\\x%02X={%s}", s.systemByte, 0x3F, "AREA")

	*list = append(*list, systemWindow, systemArea)
}
