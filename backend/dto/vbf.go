package dto

// Navegador do container .vbf (SOMENTE LEITURA): a sidebar monta a árvore a
// partir destes DTOs e só então pede o conteúdo do arquivo clicado — o
// índice inteiro do .vbf nunca é serializado para o frontend, só os filhos
// IMEDIATOS de um diretório (ListVbfDir).

// VbfRoot é uma raiz de árvore da sidebar: um .vbf encontrado perto do
// executável do jogo.
type VbfRoot struct {
	// Name é o nome do arquivo (FFX_Data.vbf / FFX2_Data.vbf).
	Name string `json:"name"`
	// Path é o caminho absoluto — é ele que as chamadas seguintes usam.
	Path string `json:"path"`
	// Version é a árvore do container (ffx | ffx2), derivada no backend.
	Version string `json:"version"`
	Size    int64  `json:"size"`
	// Entries é o total de arquivos no índice (só informação de cabeçalho).
	Entries int `json:"entries"`
}

// VbfNode é um filho imediato de um diretório do .vbf — ou um chunk do
// macrodic (filho virtual de macrodic.dcp).
type VbfNode struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  uint64 `json:"size,omitempty"`

	// Kind/ID/Version só vêm preenchidos quando o app SABE servir o
	// arquivo (o caminho bate com o que o próprio app resolveria em data/).
	// Vazio = formato fora do escopo (áudio, vídeo, executável…): o nó
	// aparece na árvore, mas clicar avisa em vez de quebrar.
	Kind    string `json:"kind,omitempty"`
	ID      string `json:"id,omitempty"`
	Version string `json:"version,omitempty"`

	// Macro marca macrodic.dcp: os filhos saem de ListVbfMacroChunks
	// (chunk_00, chunk_01, …), não de ListVbfDir.
	Macro bool `json:"macro,omitempty"`
	// Image marca uma textura: abre o painel de imagem, não a tabela.
	Image bool `json:"image,omitempty"`
}
