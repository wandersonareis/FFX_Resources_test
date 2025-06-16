package common

var (
	languageCodes = map[string]string{
		"ch": "ch",
		"kr": "kr",
		"jp": "jp",
	}

	SupportedLanguages = map[string]string{
		"ch": "Chinese",
		"de": "German",
		"fr": "French",
		"it": "Italian",
		"jp": "Japanese",
		"kr": "Korean",
		"sp": "Spanish",
		"us": "English",
	}

	Charsets = []string{"ch", "cn", "jp", "kr", "us"}
)

func LanguageCodeToCharset(languageCode string) string {
	if charset, ok := languageCodes[languageCode]; ok {
		return charset
	}
	return "us"
}

func IsSupportedLanguage(languageCode string) bool {
	_, ok := SupportedLanguages[languageCode]
	return ok
}
