package components

import "ffxresources/backend/common"

type LocalizedKeyedStringObject struct {
	contents map[string]*KeyedString
}

func NewLocalizedKeyedStringObject() *LocalizedKeyedStringObject {
	return &LocalizedKeyedStringObject{contents: make(map[string]*KeyedString)}
}

func NewLocalizedKeyedStringObjectWithContent(languageCode string, content *KeyedString) *LocalizedKeyedStringObject {
	l := NewLocalizedKeyedStringObject()
	l.SetLocalizedContent(languageCode, content)
	return l
}

func (l *LocalizedKeyedStringObject) SetLocalizedContent(languageCode string, content *KeyedString) {
	if _, ok := l.contents[languageCode]; ok && content.IsEmpty() {
		return
	} else {
		l.contents[languageCode] = content
	}
}

func (l *LocalizedKeyedStringObject) ReadAndSetLocalizedContent(languageCode string, bytes []byte, offset, key uint16) {
	if bytes == nil {
		return
	}
	charset := GetCharsetForLanguage(languageCode)
	ks := NewKeyedString(charset, offset, key, bytes)
	if ks == nil {
		return
	}
	l.SetLocalizedContent(languageCode, ks)
}

func (l *LocalizedKeyedStringObject) GetLocalizedContent(languageCode string) *KeyedString {
	return l.contents[languageCode]
}

func (l *LocalizedKeyedStringObject) GetLocalizedString(languageCode string) string {
	if ks := l.GetLocalizedContent(languageCode); ks != nil {
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
