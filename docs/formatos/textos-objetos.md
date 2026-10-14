# Textos de sistema / objetos (`battle/kernel/*.bin`)

Pacote: `backend/fileFormats/objectsfile` (motor: `objectsfile/chunkmap`,
string table: [`backend/core/strings`](../motor-strings.md))

Arquivos: `command.bin`, `item.bin`, `important.bin` (key items),
`a_ability.bin`, `w_name.bin`, `monster*.bin`, `text*.bin`, `player*.bin`,
`plate*.bin`, etc. — os "objetos" de batalha/menu do FFX e FFX-2, e os
binários do Last Mission.

## Layout do arquivo

Diferente dos [eventos](eventos.md), o formato de texto do sistema tem
**header, bloco de dados e bloco de texto separados**. O bloco de dados
tem ponteiros que apontam para os textos no bloco de texto:

```
+---------------------------+ 0x00
| Header                    |  V1: 20 bytes (FFX)
|                           |  V2: 32 bytes (FFX-2 / LastMission)
+---------------------------+ HeaderLength
| Bloco de dados (chunks)   |  N chunks de tamanho fixo
|  (um chunk = um objeto)   |  cada chunk tem ponteiros (u16 offset
|                           |  + u8 key) para a string table
+---------------------------+ HeaderLength + TotalLength
| Bloco de texto            |  string table: textos no charset do
|  (string table)           |  jogo com terminador 0x00
+---------------------------+ EOF
```

### Header

V1 (FFX, 20 bytes) — todos os campos little-endian:

```
u32 sigA @0x00    u32 sigB @0x04
u16 min @0x08     u16 max @0x0A      -- range de índices dos objetos
u16 individual @0x0C                  -- tamanho de cada chunk
u16 total @0x0E                      -- tamanho do bloco de dados
u32 padding @0x10
```

V2 (FFX-2 / LastMission, 32 bytes):

```
u32 sigA @0x00    u32 sigB @0x04    u32 unknown @0x08
u32 min @0x0C     u32 max @0x10     u32 individual @0x14
u32 total @0x18   u32 headerSize @0x1C
```

A quantidade de objetos é `max - min + 1` — **sempre** derivada do header,
nunca de `total/individual`.

### Chunks (bloco de dados)

Cada chunk é um registro de tamanho fixo (`individual` bytes) com campos
binários (stats, IDs, etc.) e **campos de texto**: pares
`u16 offset + u8 key`. O offset é **relativo ao início da string table**;
aponta para o texto do campo no bloco de texto.

Um campo de texto pode ser um segmento ponteiro simples (`u16 offset` +
`u8 key`) ou um **TextPair** — dois segmentos, o texto normal e a variante
simplificada (exportados como `Campo.Standard` e `Campo.Simplified`).

### String table (bloco de texto)

Todo o resto do arquivo depois dos chunks. Textos encodados no charset do
jogo com terminador `0x00`. Textos vazios têm offset na cauda da tabela
(após o último terminador).

## Mapeamento por struct (chunkmap)

O pacote `chunkmap` mapeia cada chunk para uma **struct Go** — uma struct
por formato de arquivo (ver `objectsfile/schema/ffx` e `schema/ffx2`).
Campos do tipo `models.Segment`/par de segmentos são texto; o resto é dado
binário preservado byte a byte.

## Funções principais (`objectsfile/chunkmap`)

- `ReadFileHeader(data, version)` — lê e valida o header (V1/V2).
- `Split(data, version)` — separa o arquivo em header, chunks e string
  table.
- `LoadFile[T](version, patternPath)` / `LoadFileFromBytes[T](...)` —
  carrega o arquivo tipado pela struct `T`; devolve
  `[]MappedTextObject` (um por chunk) + acesso à string table.
- `Decode[T](chunk)` / `StructSize[T]()` — reflete a struct `T` contra o
  chunk; `Fields[T]` lista os caminhos dos campos de texto.
- `MappedTextObject.ExportText(lang)` / `ImportText(fields)` — exporta os
  textos do objeto como `TextField`s e aplica edições ( recalcula offsets
  de TODOS os ponteiros via
  [`RebuildKeyedStrings`](../motor-strings.md)).
- `File.ToBytes()` — re-serializa o arquivo completo: header + chunks +
  string table reconstruída.
- `SchemaRegistry` / `HasChunkmapSchema(key)` — registro dos schemas por
  versão; usado por services e scripts para iterar todos os arquivos.
- `NewMappedTextObject[T](chunk, strtab, ...)` — objeto de texto de um
  chunk individual (usado por testes de roundtrip).

### `w_name.bin` (`objectsfile/weaponnames`)

- `weaponnames.Register()` — registra o schema de `w_name.bin` no
  `SchemaRegistry` (chamado no init de `backend/services`).
- `weaponnames.Collection(version, rows)` — monta a `dto.Collection` do
  w_name para os formatters.

## Garantias de save (testes)

- **Identidade**: export → import sem edição → `ToBytes` reproduz o binário
  original **byte a byte** (`checkSchema`, todos os arquivos de produção).
- **Edição de +1 caractere**: adicionar um caractere a um texto move os
  ponteiros dos textos seguintes em +1 (`TestEditingOneCharacter...`).
- **Deslocamento global**: ao editar um texto, todo ponteiro posterior
  anda pelo mesmo delta; anteriores ficam parados; ponteiros quebrados
  (MID/OOB) preservados sem virar válidos (`checkSaveKeepsSegments`).
