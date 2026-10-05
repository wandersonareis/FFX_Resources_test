package common

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// FONTE VBF: os .vbf do jogo são a fonte da verdade quando a árvore data/
// não foi (totalmente) extraída. O conteúdo dos arquivos lidos do container
// vive num OVERLAY em memória instalado por WithVbfSource — escopo por
// chamada: nada fora do escopo enxerga o overlay, e escopos se serializam
// entre si (dois .vbf abertos ao mesmo tempo não podem dividir o mapa).
//
// Dois comportamentos distintos:
//
//	SourceVbf          → SEMPRE o container (coluna Original: pristine);
//	SourceVbfPreferred → mods/ quando existe, senão o container (coluna
//	                     Traduzido/estado atual, o mesmo mods-first de
//	                     SourcePreferred trocando data/ por .vbf).
var (
	vbfMu      sync.RWMutex
	vbfOverlay map[string][]byte
	// vbfDecode cobre o que o conjunto pré-montado não previu: um caminho
	// que os readers tentem e que não esteja no overlay é decodificado sob
	// demanda, direto do .vbf (sem disco, sem cache).
	vbfDecode func(string) ([]byte, bool)
	vbfActive bool
)

// vbfKey normaliza um caminho relativo para a chave do overlay: "/" como
// separador, sem barra inicial, sem "./" e em minúsculas.
func vbfKey(p string) string {
	s := strings.ReplaceAll(p, "\\", "/")
	s = path.Clean("/" + s)
	s = strings.TrimPrefix(s, "/")
	if s == "." {
		s = ""
	}
	return strings.ToLower(s)
}

// WithVbfSource instala o overlay (caminho relativo → bytes) enquanto roda
// fn e o remove ao final, quaisquer que sejam o retorno e o pânico. Escopos
// são SERIALIZADOS: uma leitura VBF por vez.
func WithVbfSource(overlay map[string][]byte, fn func() error) error {
	return WithVbfSourceReader(overlay, nil, fn)
}

// WithVbfSourceReader é WithVbfSource com decodificador de RESERVA: todo
// caminho que fn resolver e que não esteja no overlay pré-montado passa por
// decode (que pode ler o .vbf). É o que garante que nenhuma localização
// "esquecida" no conjunto candidato degrade silenciosamente para um só
// idioma — o overlay só otimiza (decodifica uma vez em vez de toda chamada).
//
// decode pode ser nil (overlay puro, como antes).
func WithVbfSourceReader(overlay map[string][]byte, decode func(string) ([]byte, bool), fn func() error) error {
	vbfScopeMu.Lock()
	defer vbfScopeMu.Unlock()

	normalized := make(map[string][]byte, len(overlay))
	for k, v := range overlay {
		normalized[vbfKey(k)] = v
	}

	vbfMu.Lock()
	vbfOverlay = normalized
	vbfDecode = decode
	vbfActive = true
	vbfMu.Unlock()

	defer func() {
		vbfMu.Lock()
		vbfOverlay = nil
		vbfDecode = nil
		vbfActive = false
		vbfMu.Unlock()
	}()
	return fn()
}

// vbfScopeMu serializa os escopos (não protege o mapa — vbfMu faz isso).
var vbfScopeMu sync.Mutex

// VbfSourceActive informa se há um escopo VBF em curso. Chamadores com cache
// de data/ usam para NÃO memoizar conteúdo que veio do container.
func VbfSourceActive() bool {
	vbfMu.RLock()
	defer vbfMu.RUnlock()
	return vbfActive
}

// vbfLookup resolve um caminho relativo no overlay instalado. O miss cai no
// decodificador de reserva (fora do lock: ele faz I/O no .vbf).
func vbfLookup(p string) ([]byte, bool) {
	vbfMu.RLock()
	if !vbfActive {
		vbfMu.RUnlock()
		return nil, false
	}
	data, ok := vbfOverlay[vbfKey(p)]
	decode := vbfDecode
	vbfMu.RUnlock()

	if ok {
		return data, true
	}
	if decode == nil {
		return nil, false
	}
	return decode(p)
}

// vbfPathFor devolve um caminho RESOLVIDO para um arquivo do overlay. É
// absoluto-de-falsa-conta: contém caracteres ilegais em nomes de arquivo do
// Windows, então nenhum stat/leitura em disco jamais o encontra — o conteúdo
// só sai pelo FileAccessor (ReadBytes).
func vbfPathFor(p string) string {
	return filepath.Join("<vbf>", filepath.FromSlash(p))
}

// newVbfAccessor monta o accessor para as fontes de .vbf.
//
// SourceVbfPreferred mantém o mods-first habitual (arquivo traduzido em
// mods/ vence); SourceVbf lê direto do container. Caminho ausente nos dois
// casos devolve Exists=false SEM erro (o leitor degrada como em data/).
func newVbfAccessor(p string, src FileSource) (FileAccessor, error) {
	// vbf=true por padrão: a leitura NUNCA cai no disco por engano. O único
	// ramo em que ele vira false é quando mods/ prove o arquivo.
	acc := FileAccessor{RootPath: p, ResolvedPath: vbfPathFor(p), source: src, vbf: true}

	if src == SourceVbfPreferred && !DisableMods {
		modsPath := filepath.Join(GameFilesRoot, ModsFolder, filepath.FromSlash(p))
		if info, err := os.Stat(modsPath); err == nil && !info.IsDir() {
			acc.ResolvedPath = modsPath
			acc.Info = info
			acc.Size = info.Size()
			acc.Exists = true
			acc.vbf = false
			return acc, nil
		}
	}

	if data, ok := vbfLookup(p); ok {
		acc.mem = data
		acc.Size = int64(len(data))
		acc.Exists = true
	}
	return acc, nil
}

// ErrVbfMissing é devolvido ao ler um arquivo que não está no overlay.
var ErrVbfMissing = fmt.Errorf("arquivo não disponível no .vbf")
