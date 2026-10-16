// Package appmenu é a CASCA Wails do acesso externo: monta o menu nativo
// do app (Wails v3, backend puro — zero JS), coleta argumentos com dialogs
// nativos e renderiza os resultados do backend/appcore.
//
// Régua: NENHUMA regra de negócio vive aqui. Se crescer lógica, ela
// pertence ao appcore (que o CLI futuro também usa).
package appmenu

import (
	"ffxresources/backend/appcore"
)

// kinds do submenu Texto/Importar, na mesma ordem e com os MESMOS rótulos
// da árvore de data do frontend (display-names.ts — KIND_LABELS).
var textKinds = []struct {
	kind  string
	label string
}{
	{"events", "Eventos"},
	{"objects", "Sistema"},
	{"macro", "Dicionário"},
	{"lockit", "Loc Kit"},
	{"help", "Painel de ajuda"},
	{"battletext", "Textos de batalha"},
	{"cloud", "Cloud"},
	{"tutorial", "Tutorial"},
	{"menumain", "Texto menu principal"},
}

// repositoryURL é o link do código-fonte exibido no Sobre e aberto pelo
// item de menu Ajuda ▸ Código-fonte.
const repositoryURL = "https://github.com/wandersonareis/FFX_Resources_test"

// Controller conecta o menu aos dialogs nativos e às operações do appcore.
// Os handlers rodam em goroutine (o Wails v3 despacha OnClick assim), então
// dialogs e I/O são seguros sem bloquear a UI.
type Controller struct {
	// LogoUS são os bytes de docs/imagens/logo_us.png (embed no main),
	// usados pela janela Sobre.
	LogoUS []byte

	// macrodicPath guarda o dicionário APONTADO sob demanda (sessão): a
	// preparação do runtime re-aproveita nas chamadas seguintes sem
	// perguntar de novo. Vazio = ainda usa a pasta padrão.
	macrodicPath string
}

// NewController cria o controller com o logo embutido.
func NewController(logo []byte) *Controller {
	return &Controller{LogoUS: logo}
}

// appcoreFormat converte o formato do submenu para o tipo do appcore.
func appcoreFormat(json bool) appcore.ArtifactFormat {
	if json {
		return appcore.FormatJSON
	}
	return appcore.FormatStrings
}
