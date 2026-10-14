# Tabelas de evento FFX-2 (btl / cloudsave / tutorial)

Pacote: `backend/fileFormats/eventtable` (codec:
[`backend/fileFormats/event`](eventos.md))

## Layout do arquivo

Estes binários do FFX-2 usam a **MESMA tabela de strings dos eventos**
(stride 8 por entrada: `regOff u16 + flags u8 + choices u8`, `simpOff u16 +
flags u8 + choices u8`; o primeiro `u16` do arquivo é o tamanho da tabela;
offsets relativos ao início do arquivo). Ou seja: um bloco de ponteiros no
início e o bloco de texto logo depois — ver
[eventos.md](eventos.md) para o layout detalhado.

Famílias cobertas (todas com um binário por localização):

| Família | Kind | Arquivos |
|---|---|---|
| Textos de batalha | `battletext` | `battle/btl/<id>/<id>.bin` (descoberta dinâmica) |
| Cloudsave | `cloud` | `cloudsave/cloud.bin`, `cloudsave/cloudv.bin` |
| Tutorial | `tutorial` | `menu/tutorial.msb`, `menu/tuto0000.msb` |
| Menu principal (FFX) | `menumain` | `menu/menumain.bin`, `menu/menumain_btl.bin` |

## Funções principais

- `RelPath(kind, id)` — resolve o caminho do binário relativo à raiz de
  localização da versão.
- `RelPaths(kind)` / `RelPathLoc(kind, id, localization)` — variantes para
  listar família completa e caminho por idioma.
- `Key(kind, id)` / `IDs(kind)` — chave canônica do DTO e ids da família.
- `Load(kind, id, localization)` / `ReadLocalizedStringsFrom(...)` — lê o
  binário em `FieldString`s (codec de `event`).
- `event.EncodeLocalizedStrings(...)` — o codec compartilhado (header de 8
  bytes por entrada + `RebuildFieldStrings`).
- Store: `LoadFromStore`, `Get`, `List`, `Validate`, `ClearStore` — cache em
  memória dos binários carregados.

O DTO é montado pelos builders (`backend/builders`); este pacote desconhece
JSON/.strings.
