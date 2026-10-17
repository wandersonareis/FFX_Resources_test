// Artifact Writer/Reader: a camada de ARQUIVO dos formatos, padronizada
// sobre a interface Formatter (byte-level). O construtor recebe os
// parâmetros (formatter, collection, caminho, idiomas) e devolve uma
// instância PRONTA; o chamador decide QUANDO gravar/lêr e como lidar com
// os erros — Write/Read não recebem nada.
//
// A gravação/leitura é a mesma para TODOS os formatos (json, strings,
// csv, xlsx e variantes no-dedup): o byte-level é do Formatter recebido,
// tipicamente vindo do registry (Get). Naming continua decisão do
// chamador — os helpers *JSONPath resolvem o caminho canônico e SwapExt
// troca a extensão para o formato pedido.
package formatters

import (
	"fmt"
	"path/filepath"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
)

// Writer é o artefato pronto para gravar: serializa a Collection no
// formato do Formatter recebido no construtor e escreve no caminho.
type Writer interface {
	// Write serializa e grava o arquivo, devolvendo o caminho escrito.
	Write() (string, error)
}

// Reader é o artefato pronto para ler: lê o arquivo no formato do
// Formatter recebido no construtor e devolve a Collection (DTO).
type Reader interface {
	// Read lê e parseia o arquivo de volta para a Collection (DTO).
	Read() (dto.Collection, error)
}

// NewWriter constrói o gravador do artefato: recebe o formatter (registry),
// a Collection pronta, o caminho de saída e os idiomas (nil/vazio = todos).
func NewWriter(f Formatter, c dto.Collection, path string, langs []string) Writer {
	return &artifactWriter{f: f, c: c, path: path, langs: langs}
}

// NewReader constrói o leitor do artefato: recebe o formatter (registry) e
// o caminho do arquivo a parsear.
func NewReader(f Formatter, path string) Reader {
	return &artifactReader{f: f, path: path}
}

// SwapExt troca a extensão do caminho pela do formatter — o naming
// canônico é o mesmo basename para todo formato, só a extensão muda
// (no-op quando o caminho já tem a extensão do formatter).
func SwapExt(f Formatter, path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path)) + f.Extension()
}

// artifactWriter é a implementação única de Writer para todos os formatos.
type artifactWriter struct {
	f     Formatter
	c     dto.Collection
	path  string
	langs []string
}

// Write serializa (MarshalLangs com o filtro de idiomas) e grava no disco.
func (w *artifactWriter) Write() (string, error) {
	if len(w.c) == 0 {
		return "", fmt.Errorf("no text data to export")
	}
	raw, err := w.f.MarshalLangs(w.c, w.langs)
	if err != nil {
		return "", err
	}
	if err := common.WriteBytesToFile(w.path, raw); err != nil {
		return "", fmt.Errorf("error writing %s file %s: %w", w.f.Extension(), w.path, err)
	}
	common.LogVerbose("Exported %s artifact: %s", w.f.Extension(), w.path)
	return w.path, nil
}

// artifactReader é a implementação única de Reader para todos os formatos.
type artifactReader struct {
	f    Formatter
	path string
}

// Read lê o arquivo e parseia (Unmarshal) de volta para a Collection.
func (r *artifactReader) Read() (dto.Collection, error) {
	raw, err := common.ReadFile(r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s file %s: %w", r.f.Extension(), r.path, err)
	}
	c, err := r.f.Unmarshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s file %s: %w", r.f.Extension(), r.path, err)
	}
	return c, nil
}
