package tags

import (
	"fmt"
	"regexp"
)

type ICodePageGenerator interface {
	GenerateCode(byteStr string, index int) string
	GetRange() (start, end int)
}

type UCodeGenerator struct{}

func (g *UCodeGenerator) GenerateCode(byteStr string, index int) string {
	return fmt.Sprintf("%s={u%02X}", byteStr, index)
}

func (g *UCodeGenerator) GetRange() (start, end int) {
	return 0x00, 0x89
}

type XCodeGenerator struct{}

func (g *XCodeGenerator) GenerateCode(byteStr string, index int) string {
	return fmt.Sprintf("%s={x%02X}", byteStr, index)
}

func (g *XCodeGenerator) GetRange() (start, end int) {
	return 0x8A, 0xFF
}

type FFXTextTagUnknown struct {
	generators []ICodePageGenerator
}

func NewTextTagUnknown() *FFXTextTagUnknown {
	return &FFXTextTagUnknown{
		generators: []ICodePageGenerator{
			&UCodeGenerator{},
			&XCodeGenerator{},
		},
	}
}

func (u *FFXTextTagUnknown) processCodePage(codePage *[]string, generator ICodePageGenerator) {
	start, end := generator.GetRange()

	for i := start; i <= end; i++ {
		byteStr := fmt.Sprintf("\\x%02X", i)
		if u.ignoreList(i) {
			continue
		}

		if !u.codeExists(codePage, byteStr) {
			*codePage = append(*codePage, generator.GenerateCode(byteStr, i))
		}
	}
}

func (u *FFXTextTagUnknown) codeExists(codePage *[]string, byteStr string) bool {
	re := regexp.MustCompile("^" + regexp.QuoteMeta(byteStr))
	for _, v := range *codePage {
		if re.MatchString(v) {
			return true
		}
	}
	return false
}

func (u *FFXTextTagUnknown) ignoreList(value int) bool {
	return value == 0x0D || value == 0x0A
}

func (u *FFXTextTagUnknown) AddUnknownUCodePage(codePage *[]string) {
	u.processCodePage(codePage, &UCodeGenerator{})
}

func (u *FFXTextTagUnknown) AddUnknownXCodePage(codePage *[]string) {
	u.processCodePage(codePage, &XCodeGenerator{})
}
