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
		areaCodePage:         tags.NewTextTagArea(),
		buttonCodePage:       tags.NewTextTagButton(),
		characterCodePage:    tags.NewTextTagCharacter(),
		colorCodePage:        tags.NewTextTagColor(),
		itemsCodePage:        tags.NewTextItems(),
		npcCodePage:          tags.NewTextTagNPC(),
		systemCodePage:       tags.NewTextTagSystem(),
		codesCodePage:        tags.NewTextTagCode(),
		iconsCodePage:        tags.NewTextTagIcons(),
		lettersCodePage:      tags.NewLetters(),
		textCodePage:         tags.NewText(),
		unknownCodePage:      tags.NewTextTagUnknown(),
		localizationCodePage: tags.NewTextTagLocation(),
	}
}

func (e *ffxTextEncodingHelper) createFFXTextEncoding() []string {
	codePage := make([]string, 0, 520)

	codePage = slices.Concat(
		e.areaCodePage.FFXTextAreaCodePage(),
		e.buttonCodePage.FFXTextFullButtonsCodePage(),
		e.characterCodePage.FFXTextCharacterCodePage(),
		e.colorCodePage.FFXColorsCodePage(),
		e.itemsCodePage.FFXTextItemsCodePage(),
		e.npcCodePage.FFXTextNPCCodePage(),
		e.systemCodePage.FFXTextSystemCodePage(),
		e.codesCodePage.FFXTextCodesCodePage(),
		e.iconsCodePage.FFXTextIconsCodePage(),
		e.lettersCodePage.FFXTextLettersCodePage(),
		e.lettersCodePage.FFXTextSpecialLettersCodePage(),
		e.textCodePage.FFXTextTextCodePage(),
	)

	unknown := tags.NewTextTagUnknown()
	unknown.AddUnknownUCodePage(&codePage)
	unknown.AddUnknownXCodePage(&codePage)

	slices.Sort(codePage)

	return codePage
}

func (l *ffxTextEncodingHelper) createFFXTextUTF8Encoding() []string {
	return l.localizationCodePage.FFXTextUTF8CodePage()
}

func (e *ffxTextEncodingHelper) createFFXTextSimpleEncoding() []string {
	codePage := make([]string, 0, 31)

	codePage = slices.Concat(
		[]string{
			fmt.Sprintf("\\x%02X={%s}", 0x03, "NEWLINE"),
			fmt.Sprintf("\\x%02X\\c%02X={VAR%02X:\\h%02X}", 0x05, 0x01, 0x05, 0x01),
			fmt.Sprintf("\\x%02X\\c%02X={u%02X:\\h%02X}", 0x0B, 0x01, 0x0B, 0x01),
		},

		e.buttonCodePage.FFXTextButtonsCodePage(),
		e.lettersCodePage.FFXTextLettersCodePage(),
		e.lettersCodePage.FFXTextSpecialLettersCodePage(),
		e.iconsCodePage.FFXTextIconsCodePage(),
	)

	unknown := tags.NewTextTagUnknown()
	unknown.AddUnknownUCodePage(&codePage)
	unknown.AddUnknownXCodePage(&codePage)

	slices.Sort(codePage)

	return codePage
}
