package common

import "sort"

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

	Charsets = []string{"ch", "cn", "jp", "kr", "us", "default"}
)

// Language é um idioma disponível em formato chave/valor:
// Code é o código usado nos arquivos (ex: "us"), Name o nome
// de exibição (ex: "English"). Todas as versões usam os mesmos idiomas.
type Language struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// AvailableLanguages devolve os idiomas disponíveis ordenados por código.
func AvailableLanguages() []Language {
	out := make([]Language, 0, len(SupportedLanguages))
	for code, name := range SupportedLanguages {
		out = append(out, Language{Code: code, Name: name})
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Code < out[b].Code })
	return out
}

// SupportedLanguageCodes devolve só os códigos, ordenados.
// É o array a passar para extração (ids + langs) e formatters.
func SupportedLanguageCodes() []string {
	codes := make([]string, 0, len(SupportedLanguages))
	for code := range SupportedLanguages {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

// LanguageCodeToCharset resolve o charset de um idioma: ch/kr/jp têm tabela
// dedicada; sp/fr/de/it usam a tabela default (a-trema, glifo do trema
// intacto nos fonts desses idiomas); us/en/desconhecidos usam a us (a-til,
// font do us com a troca do til). Espelha ffxencoding.GetCharsetForLanguage.
func LanguageCodeToCharset(languageCode string) string {
	switch languageCode {
	case "ch", "kr", "jp":
		return languageCode
	case "sp", "fr", "de", "it":
		return "default"
	}
	return "us"
}

func IsSupportedLanguage(languageCode string) bool {
	_, ok := SupportedLanguages[languageCode]
	return ok
}
