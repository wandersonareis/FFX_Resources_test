package components

import "ffxresources/backend/common"

type LocalizedKeyedStringObject struct {
	contents map[string]*KeyedString
}

func NewLocalizedKeyedStringObject() *LocalizedKeyedStringObject {
	return &LocalizedKeyedStringObject{contents: make(map[string]*KeyedString)}
}

func NewLocalizedKeyedStringObjectWithContent(localization string, content *KeyedString) *LocalizedKeyedStringObject {
	l := NewLocalizedKeyedStringObject()
	l.SetLocalizedContent(localization, content)
	return l
}

func (l *LocalizedKeyedStringObject) SetLocalizedContent(localization string, content *KeyedString) {
	if _, ok := l.contents[localization]; ok && content.IsEmpty() {
		return
	} else {
		l.contents[localization] = content
	}
}

func (l *LocalizedKeyedStringObject) ReadAndSetLocalizedContent(localization string, bytes []byte, offset, key uint16) {
	if bytes == nil {
		return
	}
	charset := LocalizationToCharset(localization)
	ks := NewKeyedString(charset, offset, key, bytes)
	l.SetLocalizedContent(localization, ks)
}

func (l *LocalizedKeyedStringObject) GetLocalizedContent(localization string) *KeyedString {
	return l.contents[localization]
}

func (l *LocalizedKeyedStringObject) GetLocalizedString(localization string) string {
	if ks := l.GetLocalizedContent(localization); ks != nil {
		return ks.GetString()
	}
	return ""
}

func (l *LocalizedKeyedStringObject) GetDefaultContent() *KeyedString {
	return l.GetLocalizedContent(common.DefaultLocalization)
}

func (l *LocalizedKeyedStringObject) GetDefaultString() string {
	if s := l.GetLocalizedString(common.DefaultLocalization); s != "" {
		return s
	}
	return ""
}

func (l *LocalizedKeyedStringObject) CopyInto(other *LocalizedKeyedStringObject) {
	for loc, content := range l.contents {
		other.SetLocalizedContent(loc, content)
	}
}

func (l *LocalizedKeyedStringObject) String() string {
	return l.GetDefaultString()
}
