package services

import (
	"fmt"
	"path/filepath"

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

// revealInExplorer é a casca portátil do "revelar no explorador": resolve o
// caminho absoluto, registra o que foi revelado (rastreabilidade — sem isso,
// uma falha da shell aparece como "abriu a pasta errada", sem diagnóstico) e
// delega a chamada real para revealFileOnOS, implementada UMA vez por sistema,
// cada uma no seu arquivo com build tag.
//
// O runtime nunca é comutado por runtime.GOOS: um GOOS sem o seu arquivo falha
// na compilação (undefined: revealFileOnOS) em vez de rodar em silêncio a
// chamada errada.
func revealInExplorer(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("caminho inválido: %w", err)
	}
	common.LogVerbose("reveal: revelando arquivo no explorador: %s", abs)
	if err := revealFileOnOS(abs); err != nil {
		common.LogWarning("reveal: falha ao revelar %s: %v", abs, err)
		return err
	}
	return nil
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
