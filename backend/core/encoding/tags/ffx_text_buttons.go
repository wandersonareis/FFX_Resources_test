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
	return &FFXTextTagButton{
		buttonByte: 0x0B,
		buttonTag:  0x01,
		lineBreak:  LineBreakString(),
	}
}

func (b *FFXTextTagButton) FFXTextFullButtonsCodePage() []string {
	return slices.Concat(
		[]string{b.buttonCommand()},
		b.generateButtons(),
		b.buildButtonUnknownSequence(),
	)
}

func (b *FFXTextTagButton) FFXTextButtonsCodePage() []string {
	return b.generateButtons()
}

func (b *FFXTextTagButton) generateButtons() []string {
	buttonsMap := b.getButtonMap()
	buttons := make([]string, 0, len(buttonsMap))

	generateButtonCode := func(key byte, value string) string {
		return fmt.Sprintf("\\x%02X\\x%02X={%s}", b.buttonByte, key, value)
	}

	for key, value := range buttonsMap {
		buttons = append(buttons, generateButtonCode(key, value))
	}

	return buttons
}

func (b *FFXTextTagButton) buttonCommand() string {
	return fmt.Sprintf("\\x%02X\\c%02X={u%02X:\\h%02X}", b.buttonByte, b.buttonTag, b.buttonByte, b.buttonTag)
}

func (b *FFXTextTagButton) buildButtonUnknownSequence() []string {
    fBytes := []byte{0xF0, 0xF1, 0xF2, 0xF3, 0xF4, 0xF6, 0xF7, 0xF8, 0xF9, 0xFA}
    
    // Mapa que define qual unknownByte usar para cada índice
    // -1 indica que deve usar generateButtonF ao invés de generateUnknownF
    unknownByteMap := map[int]int{
        0: 0,  // 0x10
        4: 1,  // 0x08
        5: 2,  // 0x0C
        6: 3,  // 0x04
        7: 2,  // 0x0C
        9: 3,  // 0x04
    }

    unknownBytes := []byte{0x10, 0x08, 0x0C, 0x04}
    buttons := make([]string, 0, len(fBytes))

    for i, fByte := range fBytes {
        if unknownByte, exists := unknownByteMap[i]; exists {
            // Usar generateUnknownF para índices específicos
            left := fmt.Sprintf("\\x%02X\\x%02X\\c%02X", 
                b.buttonByte, fByte, unknownBytes[unknownByte])
            right := fmt.Sprintf("%s{x%02X%02X\\h%02X}", 
                b.lineBreak, b.buttonByte, fByte, unknownBytes[unknownByte])
            buttons = append(buttons, fmt.Sprintf("%s=%s", left, right))
        } else {
            // Usar generateButtonF para os demais casos
            buttons = append(buttons, b.generateButtonUnknownFormat(fByte))
        }
    }

    return buttons
}

func (b *FFXTextTagButton) generateButtonUnknownFormat(fByte byte) string {
	left := fmt.Sprintf("\\x%02X\\x%02X", b.buttonByte, fByte)
	right := fmt.Sprintf("%s{x%02X%02X}", b.lineBreak, b.buttonByte, fByte)
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