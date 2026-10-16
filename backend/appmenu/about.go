// Janela Sobre: HTML/CSS puros embutidos (WebviewWindowOptions.HTML) com o
// logo_us da documentação em base64, a descrição e o objetivo do app e o
// link do código-fonte — SEM JavaScript, como todo o resto do menu.
//
// O link do repositório aparece como TEXTO selecionável: o beta.28 do Wails
// não intercepta navegação externa (clicar num <a> levaria a própria janela
// ao site). O acesso "clicável" é o item Ajuda ▸ Código-fonte, que abre o
// navegador pelo Go (app.Browser.OpenURL).
package appmenu

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// aboutWindowName é o nome da janela Sobre (instância única).
const aboutWindowName = "sobre-ffx-resources"

// onAbout abre (ou foca) a janela Sobre.
func (c *Controller) onAbout() {
	a := app()
	if a == nil {
		return
	}
	if w, ok := a.Window.GetByName(aboutWindowName); ok {
		w.Show()
		w.Focus()
		return
	}
	a.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:      aboutWindowName,
		Title:     "Sobre",
		Width:     580,
		Height:    480,
		MinWidth:  480,
		MinHeight: 420,
		URL:       "",
		HTML:      c.aboutHTML(),
		CSS:       aboutCSS(),
		StartState: application.WindowStateNormal,
	})
}

// onSource abre o repositório no navegador padrão (Go puro, sem JS).
func (c *Controller) onSource() {
	a := app()
	if a == nil {
		return
	}
	if err := a.Browser.OpenURL(repositoryURL); err != nil {
		c.errBox("Código-fonte", "Não foi possível abrir o navegador:\n"+classifyError(err)+"\n\n"+repositoryURL)
	}
}

// aboutHTML monta a página do Sobre: logo, título, descrição, objetivo e a
// URL do código-fonte (texto selecionável).
func (c *Controller) aboutHTML() string {
	logo := ""
	if len(c.LogoUS) > 0 {
		logo = fmt.Sprintf(
			`<img class="logo" alt="FINAL FANTASY X | X-2 HD Remaster" src="data:image/png;base64,%s"/>`,
			base64.StdEncoding.EncodeToString(c.LogoUS))
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head><meta charset="utf-8"><title>Sobre</title></head>
<body>
  <main>
	%s
	<h1>FINAL FANTASY X and X-2 HD Remaster<br><small>Resources Editor</small></h1>
	<p class="desc">
	  Editor de recursos (texto e textura) dos binários de FINAL FANTASY X /
	  X-2 HD Remaster: exporta os textos do jogo para artefatos editáveis
	  (JSON e .strings), reaplica traduções nos binários e extrai / importa
	  texturas do container .dds.phyre — sempre preservando o original.
	</p>
	<p class="goal">
	  <strong>Objetivo:</strong> dar ao projeto de tradução uma ferramenta
	  local, offline e de código aberto, com validação de formato antes de
	  qualquer processamento (headers, ponteiros e lista de skip).
	</p>
	<p class="source">
	  Código-fonte:<br>
	  <span class="url">%s</span>
	</p>
	<p class="note">Construído em Go + Wails v3 — menu e dialogs 100%% nativos, sem JavaScript.</p>
  </main>
</body>
</html>`, logo, repositoryURL)
}

// aboutCSS é o estilo da página (embutido, tema escuro do app).
func aboutCSS() string {
	return strings.Join([]string{
		`body{margin:0;background:#15161a;color:#e8e8ec;font-family:"Segoe UI",system-ui,sans-serif;}`,
		`main{box-sizing:border-box;padding:28px 32px;max-width:640px;margin:0 auto;}`,
		`.logo{display:block;max-width:420px;width:100%;margin:0 auto 18px auto;image-rendering:auto;}`,
		`h1{font-size:20px;line-height:1.3;margin:0 0 14px 0;font-weight:600;text-align:center;}`,
		`h1 small{display:block;font-size:14px;font-weight:400;color:#a9a9b8;margin-top:4px;}`,
		`.desc{font-size:13.5px;line-height:1.55;color:#c9c9d4;margin:0 0 12px 0;}`,
		`.goal{font-size:13.5px;line-height:1.55;color:#c9c9d4;margin:0 0 18px 0;}`,
		`.source{font-size:13px;color:#a9a9b8;margin:0 0 6px 0;}`,
		`.url{user-select:all;color:#7aa2f7;word-break:break-all;}`,
		`.note{font-size:11.5px;color:#7c7c8a;margin-top:18px;}`,
	}, "\n")
}
