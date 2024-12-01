package ffxencoding

import (
	"ffxresources/backend/core/encoding/tags"
	"fmt"
	"slices"
)

type ffxTextEncodingHelper struct {
	areaCodePage         *tags.FFXTextTagArea
	buttonCodePage       *tags.FFXTextTagButton
	characterCodePage    *tags.FFXTextTagCharacter
	colorCodePage        *tags.FFXTextTagColor
	itemsCodePage        *tags.FFXTextItems
	npcCodePage          *tags.FFXTextTagNPC
	systemCodePage       *tags.FFXTextTagSystem
	codesCodePage        *tags.FFXTextTagCode
	iconsCodePage        *tags.FFXTextTagIcons
	lettersCodePage      *tags.FFXTextTagLetters
	textCodePage         *tags.FFXTextTagText
	unknownCodePage      *tags.FFXTextTagUnknown
	localizationCodePage *tags.FFXTextTagLocation
}

func newFFXTextEncodingHelper() *ffxTextEncodingHelper {
	return &ffxTextEncodingHelper{
		areaCodePage:      tags.NewTextTagArea(),
		buttonCodePage:    tags.NewTextTagButton(),
		characterCodePage: tags.NewTextTagCharacter(),
		colorCodePage:     tags.NewTextTagColor(),
		itemsCodePage:     tags.NewTextItems(),
		npcCodePage:       tags.NewTextTagNPC(),
		systemCodePage:    tags.NewTextTagSystem(),
		codesCodePage:     tags.NewTextTagCode(),
		iconsCodePage:     tags.NewTextTagIcons(),
		lettersCodePage:   tags.NewLetters(),
		textCodePage:      tags.NewText(),
		unknownCodePage:   tags.NewTextTagUnknown(),
	}
}

func (e *ffxTextEncodingHelper) createFFXTextEncoding() []string {
	codePage := make([]string, 0, 520)

	codePage = append(codePage, e.areaCodePage.FFXTextAreaCodePage()...)
	codePage = append(codePage, e.buttonCodePage.FFXTextFullButtonsCodePage()...)
	codePage = append(codePage, e.characterCodePage.FFXTextCharacterCodePage()...)
	codePage = append(codePage, e.colorCodePage.FFXColorsPage()...)
	codePage = append(codePage, e.itemsCodePage.FFXTextItemsCodePage()...)
	codePage = append(codePage, e.npcCodePage.FFXTextNPCCodePage()...)
	codePage = append(codePage, e.systemCodePage.FFXTextSystemCodePage()...)
	codePage = append(codePage, e.codesCodePage.FFXTextCodePage()...)
	codePage = append(codePage, e.iconsCodePage.FFXTextIconsCodePage()...)
	codePage = append(codePage, e.lettersCodePage.FFXTextLettersCodePage()...)
	codePage = append(codePage, e.lettersCodePage.FFXTextSpecialLettersCodePage()...)
	codePage = append(codePage, e.textCodePage.FFXTextTextPage()...)

	unknown := tags.NewTextTagUnknown()
	unknown.AddUnknownUCodePage(&codePage)
	unknown.AddUnknownXCodePage(&codePage)

	slices.Sort(codePage)

	seen := make(map[string]bool)
	for _, v := range codePage {
		if seen[v] {
			println("Duplicate value found:", v)
		} else {
			seen[v] = true
		}
	}

	return codePage
}

func (l *ffxTextEncodingHelper) createFFXTextLocalizationEncoding() []string {
	codePage := make([]string, 0, 31)

	codePage = append(codePage, l.localizationCodePage.FFXTextLocationPage()...)

	return codePage
}

func (e *ffxTextEncodingHelper) createFFXTextSimpleEncoding() []string {
	codePage := make([]string, 0, 31)

	codePage = append(codePage, fmt.Sprintf("\\x%02X={%s}", 0x03, "NEWLINE"))
	codePage = append(codePage, fmt.Sprintf("\\x%02X\\c%02X={VAR%02X:\\h%02X}", 0x05, 0x01, 0x05, 0x01))
	codePage = append(codePage, fmt.Sprintf("\\x%02X\\c%02X={u%02X:\\h%02X}", 0x0B, 0x01, 0x0B, 0x01))

	codePage = append(codePage, e.buttonCodePage.FFXTextButtonsCodePage()...)
	codePage = append(codePage, e.lettersCodePage.FFXTextLettersCodePage()...)
	codePage = append(codePage, e.lettersCodePage.FFXTextSpecialLettersCodePage()...)
	codePage = append(codePage, e.iconsCodePage.FFXTextIconsCodePage()...)

	unknown := tags.NewTextTagUnknown()
	unknown.AddUnknownUCodePage(&codePage)
	unknown.AddUnknownXCodePage(&codePage)

	slices.Sort(codePage)

	return codePage
}
