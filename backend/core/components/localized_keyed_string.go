package components

import (
	"ffxresources/backend/common"
	"fmt"
	"strings"
)

// LocalizedKeyedStringObject stores KeyedString values per localization
// and implements formatting and retrieval methods.
type LocalizedKeyedStringObject struct {
	contents map[string]*KeyedString
}

// NewLocalizedKeyedStringObject initializes an empty LocalizedKeyedStringObject
func NewLocalizedKeyedStringObject() *LocalizedKeyedStringObject {
	return &LocalizedKeyedStringObject{contents: make(map[string]*KeyedString)}
}

// NewLocalizedKeyedStringObjectWithContent initializes with one localization
func NewLocalizedKeyedStringObjectWithContent(localization string, content *KeyedString) *LocalizedKeyedStringObject {
	l := NewLocalizedKeyedStringObject()
	l.SetLocalizedContent(localization, content)
	return l
}

// SetLocalizedContent sets or updates the content for a localization
// but avoids overwriting existing content with an empty KeyedString
func (l *LocalizedKeyedStringObject) SetLocalizedContent(localization string, content *KeyedString) {
	if _, ok := l.contents[localization]; ok && content.IsEmpty() {
		// do not overwrite with empty
		return
	} else {
		l.contents[localization] = content
	}
}

// ReadAndSetLocalizedContent reads a KeyedString from bytes and sets it
func (l *LocalizedKeyedStringObject) ReadAndSetLocalizedContent(localization string, bytes []byte, offset, key int) {
	if bytes == nil {
		return
	}
	charset := LocalizationToCharset(localization)
	ks := NewKeyedString(charset, offset, key, bytes)
	l.SetLocalizedContent(localization, ks)
}

// WriteAllContent returns all localizations joined by newline, prefixed by display names
func (l *LocalizedKeyedStringObject) WriteAllContent() string {
	var builder strings.Builder
	for code, display := range common.Localizations {
		content := l.contents[code]
		if content != nil {
			builder.WriteString(fmt.Sprintf("[%s] %s\n", display, content.String()))
		} else {
			builder.WriteString(fmt.Sprintf("[%s] \n", display))
		}
	}
	return strings.TrimRight(builder.String(), "\n")
}

// GetLocalizedContent retrieves the KeyedString for a localization
func (l *LocalizedKeyedStringObject) GetLocalizedContent(localization string) *KeyedString {
	return l.contents[localization]
}

// GetLocalizedString retrieves the string or empty if not present
func (l *LocalizedKeyedStringObject) GetLocalizedString(localization string) string {
	if ks := l.GetLocalizedContent(localization); ks != nil {
		return ks.GetString()
	}
	return ""
}

// GetDefaultContent returns the content for the default localization
func (l *LocalizedKeyedStringObject) GetDefaultContent() *KeyedString {
	return l.GetLocalizedContent(common.DefaultLocalization)
}

// GetDefaultString returns the string for the default localization
func (l *LocalizedKeyedStringObject) GetDefaultString() string {
	if s := l.GetLocalizedString(common.DefaultLocalization); s != "" {
		return s
	}
	return ""
}

// CopyInto copies all stored contents into another LocalizedKeyedStringObject
func (l *LocalizedKeyedStringObject) CopyInto(other *LocalizedKeyedStringObject) {
	for loc, content := range l.contents {
		other.SetLocalizedContent(loc, content)
	}
}

// String implements fmt.Stringer and returns the default string
func (l *LocalizedKeyedStringObject) String() string {
	return l.GetDefaultString()
}
