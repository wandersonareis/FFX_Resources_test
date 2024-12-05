package tags

import (
	"fmt"
	"slices"
)

type FFXTextTagCode struct{}

func NewTextTagCode() *FFXTextTagCode {
	return &FFXTextTagCode{}
}

func (c *FFXTextTagCode) FFXTextCodePage() []string {
	return slices.Concat(
		c.generate2CBytesCodePage(),
		c.addVarsCode(),
	)
}

func (c *FFXTextTagCode) generate2CBytesCodePage() []string {
	get2CBytesCodeMap := c.get2CBytesCodeMap()
	byte2C := byte(0x2C)

	codePage := make([]string, 0, len(get2CBytesCodeMap))

	generate2CBytesCode := func(byte2C byte, key byte, value string) string {
		return fmt.Sprintf("\\x%02X\\x%02X={%s}", byte2C, key, value)
	}

	for key, value := range get2CBytesCodeMap {
		codePage = append(codePage, generate2CBytesCode(byte2C, key, value))
	}

	return codePage
}

func (c *FFXTextTagCode) get2CBytesCodeMap() map[byte]string {
	return map[byte]string{
		0x30: "A",
		0x34: "E",
		0x36: "G",
		0x3C: "M",
		0x3E: "O",
		0x41: "R",
		0x45: "V",
	}
}

func (c *FFXTextTagCode) addVarsCode() []string {
	codePage := make([]string, 0, 5)

	codePage = append(codePage, "\\c01={u\\h01}")
	codePage = append(codePage, "\\x07\\c01={VAR07:\\h01}")
	codePage = append(codePage, "\\x10={CHOICE???}")
	codePage = append(codePage, "\\x10\\c01={CHOICE:\\h01}")
	codePage = append(codePage, "\\x12\\c01={VAR12:\\h01}")

	return codePage
}
