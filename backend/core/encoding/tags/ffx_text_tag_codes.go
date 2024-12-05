package tags

import (
	"fmt"
	"slices"
)

type FFXTextTagCode struct {
	ffxTagsBase
}

func NewTextTagCode() *FFXTextTagCode {
	return &FFXTextTagCode{
		ffxTagsBase: ffxTagsBase{},
	}
}

func (c *FFXTextTagCode) FFXTextCodesCodePage() []string {
	return slices.Concat(
		c.addVarsCode(),
		c.generateCodesCodePage(),
	)
}

func (c *FFXTextTagCode) generateCodesCodePage() []string {
	return c.processCodePage(&ffxCodes{codesID: 0x2C})
}

func (c *FFXTextTagCode) addVarsCode() []string {
	return []string{
		"\\c01={u\\h01}",
		"\\x07\\c01={VAR07:\\h01}",
		"\\x10={CHOICE???}",
		"\\x10\\c01={CHOICE:\\h01}",
		"\\x12\\c01={VAR12:\\h01}",
	}
}

type ffxCodes struct {
	codesID byte
}

func (c *ffxCodes) getMap() map[byte]string {
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

func (c *ffxCodes) generateCode(key byte, value string) string {
	return fmt.Sprintf("\\x%02X\\x%02X={%s}", c.codesID, key, value)
}
