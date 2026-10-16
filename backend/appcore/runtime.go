// Runtime de texto do acesso externo: o que uma operação de TEXTO precisa
// carregar ANTES de decodificar/aplicar binário. É a única dependência do
// appcore que ainda toca na árvore do jogo (charsets da versão e dicionário
// de macros na localização padrão).
//
// TODO: deixar a necessidade GRANULAR de encoding e macrodic.
//
//	Hoje prepareTextRuntime carrega encoding + macrodic em bloco "por
//	segurança", mas a necessidade real é granular:
//
//	  - imagem (.dds.phyre) NÃO precisa de encoding;
//	  - lockit precisa apenas de encoding (charsets);
//	  - o restante (events, objects, macro, help, tabelas) precisa do
//	    macrodic — as tags MCR do texto resolvem pelo dicionário
//	    (datastore.GetMacro, publicado por PublishStrings).
//
//	Aqui estamos presumindo que o macrodic SEMPRE existe e está ao alcance
//	(pasta padrão menu/macrodic.dcp da localização, na árvore do jogo).
//	No uso do menu ou do CLI isso pode NÃO ser verdade — o usuário aponta
//	binários avulsos sem árvore extraída. Será necessário:
//
//	  - no MENU: apontar sob demanda (dialog pedindo o macrodic.dcp quando
//	    o ErrMacrodicUnavailable abaixo subir) — o gancho já existe:
//	    Request.MacrodicPath;
//	  - no CLI: detectar ou exigir flag (--macrodic) apontando o arquivo.
//
//	Quando a granularidade acima for implementada, lockit/imagem deixam de
//	pagar o custo (e a dependência) do macrodic, e o macrodic avulso passa a
//	ser publicado a partir do caminho apontado.
package appcore

import (
	"fmt"
	"os"
	"path/filepath"

	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/services"
)

// TextRuntime é o estado preparado para operar texto de uma versão:
// charsets carregados e dicionário de macros publicado no datastore.
type TextRuntime struct {
	Version common.GameVersion

	// MacrodicSource registra de onde o dicionário veio ("árvore padrão"
	// ou o caminho apontado sob demanda) — informativo para logs.
	MacrodicSource string
}

// prepareTextRuntime carrega o que o kind pede. kind é o kind lógico
// (services.Kind*); macrodicPath aponta o dicionário quando o chamador já
// o tem (sob demanda do menu/CLI); "" usa a pasta padrão.
//
// Falha do macrodic no lugar padrão vira ErrMacrodicUnavailable (o chamador
// aponta e repete); falha de charset é erro comum (o app não funciona sem).
func prepareTextRuntime(version common.GameVersion, kind string, macrodicPath string) (*TextRuntime, error) {
	rt := &TextRuntime{Version: version}

	// Encoding: todo kind de texto decodifica texto — os charsets são
	// EMBUTIDOS (charset_tables.go), a carga é idempotente e sem I/O.
	// PrepareVersionCharsets publica os mapas; EnsureAllCharsetsLoaded
	// (usado pelos loads do app) só confere.
	if err := ffxencoding.PrepareVersionCharsets(version); err != nil {
		return nil, fmt.Errorf("charsets da versão %s: %w", version, err)
	}
	if err := ffxencoding.EnsureAllCharsetsLoaded(version); err != nil {
		return nil, fmt.Errorf("charsets da versão %s: %w", version, err)
	}

	// TODO (granular): lockit precisa apenas de encoding — desativar o
	// bloco de macrodic abaixo para ele quando a granularidade cair.
	if kind == services.KindLockit {
		return rt, nil
	}

	// Macrodic: o restante resolve tags MCR pelo dicionário. Primeiro o
	// caminho apontado (menu/CLI sob demanda); sem ele, a pasta padrão.
	if macrodicPath != "" {
		if err := publishMacrodicFromPath(macrodicPath, version); err != nil {
			return nil, err
		}
		rt.MacrodicSource = macrodicPath
		return rt, nil
	}

	containers, err := macrodic.ReadMacroDictionaryContainers(version)
	if err != nil || len(containers) == 0 {
		// Presunção quebrada (ver TODO do pacote): o macrodic pode não
		// estar na pasta padrão. O chamador aponta e repete.
		return nil, ErrMacrodicUnavailable
	}
	for _, c := range containers {
		if err := c.PublishStrings(); err != nil {
			return nil, fmt.Errorf("publicando macros: %w", err)
		}
	}
	rt.MacrodicSource = "árvore padrão"
	return rt, nil
}

// publishMacrodicFromPath publica no datastore o dicionário APONTADO pelo
// usuário (menu sob demanda / CLI). O arquivo é um macrodic.dcp da
// localização padrão (us) — a que importa para a resolução de MCR.
func publishMacrodicFromPath(path string, version common.GameVersion) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("lendo macrodic apontado: %w", err)
	}
	if filepath.Ext(path) != ".dcp" {
		return Incompatible(path, "dicionário de macros", "", "extensão esperada: .dcp")
	}
	c, err := macrodic.NewMacroDictionaryBinaryFileFromBytes(data, common.DefaultLocalization, version)
	if err != nil {
		return Incompatible(path, "dicionário de macros", "", "header de offsets de chunks não reconhecido")
	}
	if err := c.PublishStrings(); err != nil {
		return fmt.Errorf("publicando macros de %s: %w", filepath.Base(path), err)
	}
	return nil
}
