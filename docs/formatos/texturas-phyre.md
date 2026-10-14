# Texturas PhyreEngine (`.dds.phyre`)

Pacote: `backend/fileFormats/ddsphyre`

Texturas DX11 do FFX HD Remaster (PhyreEngine), extraídas para DDS/PNG e
reempacotáveis.

## Layout do arquivo

O container é **auto-descritivo**: o bloco *namespace* declara classes,
membros e a tabela de strings, de onde saem os offsets de
width/height/mips — nada é hardcoded. O payload a partir de `dataOffset`
já é um arquivo DDS:

```
u32 magic 0x50485952  -- little-endian ("RYHP" no disco;
u32 platformId         arquivo big-endian "PHYR" é rejeitado)
       0x44583131     -- "DX11"
u32 size == 84         -- 21 campos u32
u32 namespaceSize      -- deve caber no arquivo
⟨namespace: classes, membros, tabela de strings⟩
⟨payload DDS (a partir de dataOffset)⟩
```

Regras de campo (port fiel de dds-phyre-tool):

- magic little-endian `0x50485952`; big-endian rejeitado.
- platformId `0x44583131` ("DX11" no disco).
- `size == 84` (21 campos `u32`).

Extrair = header DDS + cópia bruta do payload; reempacotar = reescrever o
payload e os campos que ele altera (width/height/mips/datasize).

## Funções principais

### Paths / descoberta

- `Scan(version)` — varre a árvore de texturas (ids, somente-mods,
  fallback).
- `RootForVersion` / `RelPath` / `ExportPaths` / `ValidID` / `Exists` —
  resolução de caminhos por versão e id.

### Container

- (phyre.go) leitura/escrita do container: parse do namespace, extração do
  payload DDS, reempacotamento.
- `namespace.go` — parse do bloco namespace: `className`, `findClass`,
  `findMember`, `typeName` (offsets de width/height/mips saem daqui).

### DDS / PNG

- `dds.go` — header DDS (`prepareDDSHeader(format, width, height, mips)`,
  `parseDDSHeader`, `bufferSizeByFormat`).
- `png.go` — conversão DDS↔PNG (com flip de orientação).

### Dedupe / índice

- `BuildIndex(version)` / `IndexFor(version)` — índice de cópias por
  payload; `PayloadHash` — hash do payload.
- `Index.Copies(id)` / `Index.OriginalGroup(id)` — grupos de texturas com
  payload idêntico (o jogo reutiliza texturas).
- `InvalidateIndex(version)` / `InvalidateIndexes()` — invalidação de
  cache após edição.
