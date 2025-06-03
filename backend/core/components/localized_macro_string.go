package components

import "ffxresources/backend/common"

type LocalizedMacroStringObject struct {
	contents map[string]*MacroString
}

// Construtor
func NewLocalizedMacroStringObject() *LocalizedMacroStringObject {
	return &LocalizedMacroStringObject{
		contents: make(map[string]*MacroString),
	}
}

func NewWithContent(localization string, content *MacroString) *LocalizedMacroStringObject {
	obj := NewLocalizedMacroStringObject()
	obj.SetLocalizedContent(localization, content)
	return obj
}

func (l *LocalizedMacroStringObject) SetLocalizedContent(localization string, content *MacroString) {
	if existing, ok := l.contents[localization]; ok && content.IsEmpty() && !existing.IsEmpty() {
		return
	}
	l.contents[localization] = content
}

func (l *LocalizedMacroStringObject) GetLocalizedContent(localization string) *MacroString {
	return l.contents[localization]
}

func (l *LocalizedMacroStringObject) GetLocalizedString(localization string) string {
	if obj := l.GetLocalizedContent(localization); obj != nil {
		return obj.GetString()
	}
	return ""
}

func (l *LocalizedMacroStringObject) GetDefaultContent() *MacroString {
	return l.GetLocalizedContent(common.DefaultLocalization)
}

func (l *LocalizedMacroStringObject) CopyInto(other *LocalizedMacroStringObject) {
	for k, v := range l.contents {
		other.SetLocalizedContent(k, v) // aviso: compartilha referência
	}
}

func (l *LocalizedMacroStringObject) String() string {
	if def := l.GetDefaultContent(); def != nil {
		return def.String()
	}
	return ""
}
