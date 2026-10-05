package services

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	"ffxresources/backend/common"
)

// RevealEntryFile abre o explorador de arquivos do sistema já com o
// ARQUIVO DA ENTRADA selecionado ("Abrir até o arquivo" do menu de
// contexto).
//
// O caminho é resolvido exatamente como o app lê (mods-first, com queda
// para data/): o usuário enxerga o arquivo que está VALENDO — a substituição
// em mods/ quando existe, senão o pristine em data/. Servir data/ num caso
// em que o jogo carrega mods/ mandaria o usuário no caminho errado.
func (s *MetadataService) RevealEntryFile(kind, id string, version common.GameVersion) error {
	if err := ensureVersionReady(version); err != nil {
		return err
	}
	abs, err := revealTargetPath(kind, id, version)
	if err != nil {
		return err
	}
	return revealInExplorer(abs)
}

// revealTargetPath devolve o caminho absoluto que o "Abrir até o arquivo"
// deve revelar: mods-first como o app lê, com queda para data/ (e erro
// claro quando o arquivo não existe em lugar nenhum — abrir uma pasta sem o
// item seria enganoso).
func revealTargetPath(kind, id string, version common.GameVersion) (string, error) {
	rel, ok := originalRelPath(kind, id, version)
	if !ok {
		return "", fmt.Errorf("%s: sem arquivo associado", id)
	}
	acc, err := common.NewFileAccessorFrom(rel, common.SourceMods)
	if err != nil {
		return "", err
	}
	if !acc.Exists {
		if acc, err = common.NewFileAccessorFrom(rel, common.SourceData); err != nil {
			return "", err
		}
		if !acc.Exists {
			return "", fmt.Errorf("%s: arquivo não encontrado em mods/ nem em data/", id)
		}
	}
	return acc.ResolvedPath, nil
}

// revealInExplorer abre a pasta contendo path com o arquivo selecionado.
//
// Windows: `explorer.exe /select,"<path>"` é o modo de FOCAR um item —
// `explorer <pasta>` só abriria a pasta e o usuário teria de procurar.
// Não se faz Wait(): o explorer pode devolver exit code 1 depois de abrir a
// janela, e o app não deve ficar preso junto com o explorador.
func revealInExplorer(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("caminho inválido: %w", err)
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer.exe", `/select,"`+abs+`"`)
	case "darwin":
		cmd = exec.Command("open", "-R", abs)
	default:
		cmd = exec.Command("xdg-open", filepath.Dir(abs))
	}
	return cmd.Start()
}
