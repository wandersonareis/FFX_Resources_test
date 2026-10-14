# Dicionário de macros (`macrodic.dcp`)

Pacote: `backend/fileFormats/macrodic`

Arquivo: `menu/macrodic.dcp` — uma variante por localização. O dicionário
agrega TODAS as localizações como um único dicionário lógico no datastore
(chave de macro = `chunk*0x100 + índice`).

## Layout do arquivo

```
+------------------------------------+ 0x00
| Tabela de offsets de chunks        |  16 x u32 (64 bytes)
|  (terminador 0xFFFFFFFF opcional)  |  cada offset aponta para o
|                                    |  início de um chunk
+------------------------------------+ @chunkOffset[i]
| Chunks                             |  sequências de segmentos de
|                                    |  texto (macros), encodadas no
|                                    |  charset do idioma
+------------------------------------+ EOF
```

- O header é a tabela de 16 offsets (`u32` cada), um por chunk.
- Cada chunk segue a MESMA estrutura das tabelas de evento, em miniatura:
  um bloco de ponteiros no início (pares `u16 nameOffset + u16
  simplifiedOffset`, 4 bytes por macro; a contagem vem de
  `firstOffset / 4`) seguido do bloco de texto do chunk. Offsets iguais =
  não há variante simplificada distinta.
- Macros são textos curtos com placeholders de escolha referenciados pelos
  textos de eventos/objetos como `{MCR:...}`.

## Funções principais

### Leitura

- `NewMacroDictionaryBinaryFile(localization, version)` — handle de uma
  localização; `LoadFromBinary` popula do disco.
- `MapChunks()` — offsets dos chunks parseados dos bytes.
- `MapStrings(chunkIndex)` / `MapAllStrings()` /
  `MapStringsWithLocalization(...)` — mapeiam chunks em `[]*MacroString`.
- `ReadMacroDictionaryContainers(version)` /
  `ReadMacroDictionaryContainersFrom(...)` — lê todas as localizações de
  uma vez.

### Escrita / rebuild

- `RebuildChunkData(...)` — reconstrói os bytes de um chunk a partir das
  macros editadas (recalcula offsets internos).
- `BuildContainerBinary(...)` — re-serializa o arquivo completo (tabela de
  offsets + chunks).
- `SaveMacroDictionaryBinaries(...)` — grava em `mods/<loc>/menu/macrodic.dcp`.

### Publicação no datastore

- `PublishStrings()` / `PublishMacroDictionaryContainers(...)` — mescla os
  chunks de cada localização no datastore versionado. Objetos são
  compartilhados por ID entre localizações: a default publica primeiro e
  as demais acrescentam conteúdo localizado.
- `MaxIndex()`, `DefaultFirstLocalizations()`, `MacroRelPath(...)`,
  `SortedLocalizationKeys()` — utilidades de leitura/ordenação.

### Texto

- `NewLocalizedMacroStringObject(...)` / `MacroDictionaryTextFile` — ciclo
  de texto (export/import JSON) de uma localização.

## Relação com outros formatos

As macros são consumidas pelo jogo dentro dos textos de
[eventos](eventos.md), [objetos](textos-objetos.md) e
[ajuda](ajuda-sps2.md) via lookup `{MCR:...}` no
datastore (`datastore.GetMacro(gameVersion, id)`).
