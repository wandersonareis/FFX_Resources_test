package tags

import (
	"fmt"
	"slices"
)

type FFXTextTagButton struct {
	buttonByte byte
	buttonTag  byte
	lineBreak  string
}

func NewTextTagButton() *FFXTextTagButton {
	buttonByte := byte(0x0B)
	buttonTag := byte(0x01)

	return &FFXTextTagButton{
		buttonByte: buttonByte,
		buttonTag:  buttonTag,
		lineBreak:  LineBreakString(),
	}
}

func (b *FFXTextTagButton) FFXTextFullButtonsCodePage() []string {
	buttons := make([]string, 0, len(b.getButtonMap())+1)

	buttons = append(buttons, b.buttonCommand())

	b.generateButtonsCodePage(&buttons)

	buttons = slices.Concat(buttons, b.generateButtonsF())

	return buttons
}

func (b *FFXTextTagButton) FFXTextButtonsCodePage() []string {
	buttons := make([]string, 0, len(b.getButtonMap()))

	b.generateButtonsCodePage(&buttons)

	return buttons
}

func (b *FFXTextTagButton) generateButtonsCodePage(codePage *[]string) {
	buttonsMap := b.getButtonMap()

	for key, value := range buttonsMap {
		*codePage = append(*codePage, b.generateButtonCode(key, value))
	}
}

func (b *FFXTextTagButton) generateButtonCode(key byte, value string) string {
	return fmt.Sprintf("\\x%02X\\x%02X={%s}", b.buttonByte, key, value)
}

func (b *FFXTextTagButton) buttonCommand() string {
	return fmt.Sprintf("\\x%02X\\c%02X={u%02X:\\h%02X}", b.buttonByte, b.buttonTag, b.buttonByte, b.buttonTag)
}

func (b *FFXTextTagButton) generateButtonsF() []string {
	fBytes := []byte{0xF0, 0xF1, 0xF2, 0xF3, 0xF4, 0xF6, 0xF7, 0xF8, 0xF9, 0xFA}
	unknownBytes := []byte{0x10, 0x08, 0x0C, 0x04}

	buttonFunctions := []func(byte) string{
		func(fByte byte) string { return b.generateUnknownF(fByte, unknownBytes[0]) }, // 0
		b.generateButtonF, // 1
		b.generateButtonF, // 2
		b.generateButtonF, // 3
		func(fByte byte) string { return b.generateUnknownF(fByte, unknownBytes[1]) }, // 4
		func(fByte byte) string { return b.generateUnknownF(fByte, unknownBytes[2]) }, // 5
		func(fByte byte) string { return b.generateUnknownF(fByte, unknownBytes[3]) }, // 6
		func(fByte byte) string { return b.generateUnknownF(fByte, unknownBytes[2]) }, // 7
		b.generateButtonF, // 8
		func(fByte byte) string { return b.generateUnknownF(fByte, unknownBytes[3]) }, // 9
	}

	buttons := make([]string, 0, len(fBytes))

	for i, fByte := range fBytes {
		if i < len(buttonFunctions) {
			buttons = append(buttons, buttonFunctions[i](fByte))
		}
	}

	return buttons
}

func (b *FFXTextTagButton) generateButtonF(fByte byte) string {
	left := fmt.Sprintf("\\x%02X\\x%02X", b.buttonByte, fByte)
	right := fmt.Sprintf("%s{x%02X%02X}", b.lineBreak, b.buttonByte, fByte)
	return fmt.Sprintf("%s=%s", left, right)
}

func (b *FFXTextTagButton) generateUnknownF(fByte, unknownByte byte) string {
	left := fmt.Sprintf("\\x%02X\\x%02X\\c%02X", b.buttonByte, fByte, unknownByte)
	right := fmt.Sprintf("%s{x%02X%02X\\h%02X}", b.lineBreak, b.buttonByte, fByte, unknownByte)
	return fmt.Sprintf("%s=%s", left, right)
}

func (b *FFXTextTagButton) getButtonMap() map[byte]string {
	buttonMap := map[byte]string{
		0x30: "Button Triangle",
		0x31: "Button Circle",
		0x32: "Button X",
		0x33: "Button Square",
		0x34: "Button L1",
		0x35: "Button R1",
		0x36: "Button L2",
		0x37: "Button R2",
		0x38: "Button START",
		0x39: "Button SELECT",
		0x40: "Direcional",
		0x41: "Direcional Up",
		0x42: "Direcional Right",
		0x43: "Direcional Up+Right",
		0x44: "Direcional Down",
		0x45: "Direcional Up+Down",
		0x46: "Direcional Down+Right",
		0x47: "Direcional Up+Right+Down",
		0x48: "Direcional Left",
		0x49: "Direcional Up+Left",
		0x4A: "Direcional Left+Right",
		0x4B: "Direcional Up+Left+Right",
		0x4C: "Direcional Left+Down",
		0x4D: "Direcional Up+Left+Down",
		0x4E: "Direcional Left+Down+Right",
		0x4F: "Direcional All",
	}
	return buttonMap
}