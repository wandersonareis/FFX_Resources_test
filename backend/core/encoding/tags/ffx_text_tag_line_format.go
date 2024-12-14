package tags

import (
	"fmt"
	"slices"
)

type FFXTextTagText struct {
	ffxTagsBase
}

func NewText() *FFXTextTagText {
	return &FFXTextTagText{
		ffxTagsBase: ffxTagsBase{},
	}
}

func (t *FFXTextTagText) FFXTextTextCodePage() []string {
	return slices.Concat(
		t.generateLineBreaks(),
		t.generateLineFormat(),
	)
}

func (t *FFXTextTagText) generateLineBreaks() []string {
	return t.processCodePage(&ffxLineBreaks{})
}

func (t *FFXTextTagText) generateLineFormat() []string {
	return t.processCodePage(&ffxLineFormat{lineFormatByte: 0x0E})
}

type ffxLineBreaks struct{}

func (t *ffxLineBreaks) generateCode(key byte, value string) string {
	return fmt.Sprintf("\\x%02X=%s", key, value)
}

func (t *ffxLineBreaks) getMap() map[byte]string {
	return map[byte]string{
		0x00: "{END}",
		0x01: fmt.Sprintf("%s{NEWPAGE}%s", LineBreakString(), LineBreakString()),
		0x03: LineBreakString(),
	}
}

type ffxLineFormat struct {
	lineFormatByte byte
}

func (f *ffxLineFormat) generateCode(key byte, value string) string {
	return fmt.Sprintf("\\x%02X\\x%02X={%s}", f.lineFormatByte, key, value)
}

func (t *ffxLineFormat) getMap() map[byte]string {
	return map[byte]string{
		0x40: "Font Italic",
		0x41: "Font Normal",
	}
}
