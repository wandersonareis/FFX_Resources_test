package tags

import "fmt"

type FFXTextTagCode struct{}

func NewTextTagCode() *FFXTextTagCode {
	return &FFXTextTagCode{}
}

func (c *FFXTextTagCode) FFXTextCodePage() []string {
	codePage := make([]string, 0, 8)

	c.generate2CBytesCodePage(&codePage)
	c.addVarsCode(&codePage)

	return codePage
}

func (c *FFXTextTagCode) generate2CBytesCodePage(codePage *[]string) {
	bytesMap := c.get2CBytesCodeList()
	byte2C := byte(0x2C)

	for key, value := range bytesMap {
		*codePage = append(*codePage, c.generate2CBytesCode(byte2C, key, value))
	}
}

func (c *FFXTextTagCode) generate2CBytesCode(byte2C byte, key byte, value string) string {
	return fmt.Sprintf("\\x%02X\\x%02X={%s}", byte2C, key, value)
}

func (c *FFXTextTagCode) get2CBytesCodeList() map[byte]string {
	bytesMap := map[byte]string{
		0x30: "A",
		0x34: "E",
		0x36: "G",
		0x3C: "M",
		0x3E: "O",
		0x41: "R",
		0x45: "V",
	}

	return bytesMap
}

func (c *FFXTextTagCode) addVarsCode(codePage *[]string) {
	*codePage = append(*codePage, "\\c01={u\\h01}")
	*codePage = append(*codePage, "\\x07\\c01={VAR07:\\h01}")
	*codePage = append(*codePage, "\\x10={CHOICE???}")
	*codePage = append(*codePage, "\\x10\\c01={CHOICE:\\h01}")
	*codePage = append(*codePage, "\\x12\\c01={VAR12:\\h01}")
}
