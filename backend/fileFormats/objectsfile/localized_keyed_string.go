package objectsfile

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

type (
	ILocalizedTextObject interface {
		GetName(languageCode string) string
		GetKeyedString(title string) *LocalizedKeyedStringObject
		GetLocalizedKeyedStrings(languageCode string) []*KeyedString
		SetLocalizations(other components.LocalizationSetter)
		GetTextObject() ILocalizedTextObject
		GetHeaderLength() int
		ToBytes(languageCode string) ([]byte, error)
		ToString(languageCode string) string
		String() string
	}
	LocalizedKeyedStringObject struct {
		contents map[string]datastore.IGlobalKeyedString
	}
)

func NewLocalizedKeyedStringObject() datastore.IGlobalLocalizedKeyedStringObject {
	return &LocalizedKeyedStringObject{contents: make(map[string]datastore.IGlobalKeyedString)}
}

func NewLocalizedKeyedStringObjectWithContent(languageCode string, content *KeyedString) datastore.IGlobalLocalizedKeyedStringObject {
	l := NewLocalizedKeyedStringObject()
	l.SetLocalizedContent(languageCode, content)
	return l
}

func (l *LocalizedKeyedStringObject) SetLocalizedContent(languageCode string, content datastore.IGlobalKeyedString) {
	if _, ok := l.contents[languageCode]; ok && content.IsEmpty() {
		return
	} else {
		l.contents[languageCode] = content
	}
}

func (l *LocalizedKeyedStringObject) ReadAndSetLocalizedContent(languageCode string, bytes []byte, offset models.Offset, key models.Key, version int) {
	if bytes == nil {
		return
	}
	charset := ffxencoding.GetCharsetForLanguage(languageCode)
	ks := NewKeyedString(charset, models.Segment{Offset: offset, Key: key}, bytes, version)
	if ks == nil {
		return
	}
	l.SetLocalizedContent(languageCode, ks)
}

func (l *LocalizedKeyedStringObject) GetLocalizedContent(languageCode string) datastore.IGlobalKeyedString {
	return l.contents[languageCode]
}

func (l *LocalizedKeyedStringObject) GetLocalizedString(languageCode string) string {
	if ks := l.GetLocalizedContent(languageCode); ks != nil {
		return ks.GetString()
	}
	return ""
}

func (l *LocalizedKeyedStringObject) GetDefaultContent() datastore.IGlobalKeyedString {
	return l.GetLocalizedContent(common.DefaultLocalization)
}

func (l *LocalizedKeyedStringObject) GetDefaultString() string {
	if s := l.GetLocalizedString(common.DefaultLocalization); s != "" {
		return s
	}
	return ""
}

func (l *LocalizedKeyedStringObject) CopyInto(other datastore.IGlobalLocalizedKeyedStringObject) {
	for loc, content := range l.contents {
		other.SetLocalizedContent(loc, content)
	}
}

func (l *LocalizedKeyedStringObject) String() string {
	return l.GetDefaultString()
}
