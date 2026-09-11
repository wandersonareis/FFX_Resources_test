package datastore

import (
	"bytes"
	"ffxresources/backend/core/components"
	"ffxresources/backend/models"
)

// Interfaces genéricas para evitar dependências circulares
type INameableObject interface {
	GetName(localization string) string
}

type ILocalizedTextObject interface {
	GetLocalizedContent(localization string) IStringContent
	GetName(localization string) string // Adicionado este método
}

type IStringContent interface {
	String() string
	IsEmpty() bool
}

type IMacroObject interface {
	GetLocalizedContent(localization string) IStringContent
}

// IEventObject interface para manter compatibilidade
type IEventObject interface {
	GetName() string
	GetID() string
}

// Interface global compatível com objectsfile.ILocalizedTextObject
// Não depende de tipos concretos do pacote objectsfile

type IGlobalLocalizedKeyedStringObject interface {
	ReadAndSetLocalizedContent(languageCode string, bytes []byte, offset models.Offset, key models.Key, version int)
	SetLocalizedContent(languageCode string, content IGlobalKeyedString)
	GetLocalizedContent(languageCode string) IGlobalKeyedString
	GetLocalizedString(languageCode string) string
	GetDefaultContent() IGlobalKeyedString
	GetDefaultString() string
	CopyInto(other IGlobalLocalizedKeyedStringObject)
	String() string
}

type IGlobalKeyedString interface {
	GetOffset() models.Offset
	SetOffset(offset models.Offset)
	SetKey(key models.Key)
	GetKey() models.Key
	SetHeaderBytes(buf *bytes.Buffer)
	GetHeaderBytes(buf *bytes.Buffer)
	SetCharset(charset string)
	GetCharset() string
	GetString() string
	SetString(s, charset string)
	IsEmpty() bool
	String() string
}

type IGlobalLocalizationSetter interface {
	SetLocalizations(other IGlobalLocalizationSetter)
}

type IGlobalLocalizedTextObject interface {
	GetName(languageCode string) string
	GetKeyedString(title string) IGlobalLocalizedKeyedStringObject
	GetLocalizedKeyedStrings(languageCode string) []IGlobalKeyedString
	SetLocalizations(other IGlobalLocalizationSetter)
	GetTextObject() IGlobalLocalizedTextObject
	GetHeaderLength() int
	ToBytes(languageCode string) ([]byte, error)
	ToString(languageCode string) string
	String() string
}

// Interface global compatível com macrodic.LocalizedMacroStringObject
type IGlobalLocalizedMacroStringObject interface {
	SetLocalizedContent(localization string, content IGlobalMacroString)
	GetLocalizedContent(localization string) (IGlobalMacroString, bool)
	GetLocalizedString(localization string) string
	GetDefaultContent() (IGlobalMacroString, bool)
	CopyInto(other IGlobalLocalizedMacroStringObject)
	String() string
}

// Interface global compatível com macrodic.MacroString
type IGlobalMacroString interface {
	GetString() string
	IsEmpty() bool
	String() string
}

// IBinaryFile orquestra todo o ciclo de vida de um arquivo binário de localização
type IBinaryFile interface {
	LoadFromBinary() error
	ExportToJson(filePath string) error
	ImportFromJson(filePath string) error
	SaveToBinary(filePath string) error
	GetObjects() components.IList[IGlobalLocalizedTextObject]
}
