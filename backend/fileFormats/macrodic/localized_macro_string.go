package macrodic

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
)

type LocalizedMacroStringObject struct {
	contents components.IMap[string, datastore.IGlobalMacroString]
}

func NewLocalizedMacroStringObject() *LocalizedMacroStringObject {
	return &LocalizedMacroStringObject{
		contents: components.NewEmptyMap[string, datastore.IGlobalMacroString](),
	}
}

func NewWithContent(localization string, content *MacroString) *LocalizedMacroStringObject {
	obj := NewLocalizedMacroStringObject()
	obj.SetLocalizedContent(localization, content)
	return obj
}

func (l *LocalizedMacroStringObject) SetLocalizedContent(localization string, content datastore.IGlobalMacroString) {
	if existing, ok := l.contents.Get(localization); ok && content.IsEmpty() && !existing.IsEmpty() {
		return
	}
	l.contents.Add(localization, content)
}

func (l *LocalizedMacroStringObject) GetLocalizedContent(localization string) (datastore.IGlobalMacroString, bool) {
	return l.contents.Get(localization)
}

func (l *LocalizedMacroStringObject) GetLocalizedString(localization string) string {
	if obj, exists := l.GetLocalizedContent(localization); exists && obj != nil {
		return obj.GetString()
	}
	return ""
}

func (l *LocalizedMacroStringObject) GetDefaultContent() (datastore.IGlobalMacroString, bool) {
	return l.GetLocalizedContent(common.DefaultLocalization)
}

func (l *LocalizedMacroStringObject) CopyInto(other datastore.IGlobalLocalizedMacroStringObject) {
	/* for k, v := range l.contents {
		other.SetLocalizedContent(k, v)
	} */
	l.contents.ForEach(func(key string, value datastore.IGlobalMacroString) {
		if value != nil {
			other.SetLocalizedContent(key, value)
		}
	})
}

func (l *LocalizedMacroStringObject) String() string {
	if content, exists := l.GetDefaultContent(); exists && content != nil {
		return content.String()
	}
	return ""
}
