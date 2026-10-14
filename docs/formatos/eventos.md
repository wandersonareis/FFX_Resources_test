# Eventos (`event/obj_ps3/*.bin`)

Pacote: `backend/fileFormats/event`

## Layout do arquivo

Cada arquivo de evento contém **um bloco de ponteiros para os textos e um
bloco de texto no mesmo arquivo**, sem header global:

```
+-----------------------------+ offset 0
| Tabela de ponteiros         |  N entradas de 8 bytes
|  (8 bytes por FieldString)  |
+-----------------------------+ offset N*8 (= início do bloco de texto)
| Bloco de texto              |  textos encodados no charset do jogo,
|  (strings com terminador 0) |  comprimidos por dedup
+-----------------------------+
```

Cada entrada de 8 bytes (`FieldString`) tem **dois ponteiros** — o texto
normal e a variante simplificada (ideogramas):


```
u16 regularOffset   -- offset do texto no bloco de texto (relativo ao
                     início do ARQUIVO, não da tabela)
u8  regularFlags
u8  regularChoices  -- nº de escolhas de menu embutidas no texto
u16 simplifiedOffset
u8  simplifiedFlags
u8  simplifiedChoices
```

- `regularOffset == simplifiedOffset` significa que não há variante
  simplificada distinta.
- O primeiro `u16` do arquivo é sempre `N*8` — o leitor usa isso para
  contar as entradas (`first % 8 == 0` valida o arquivo).
- Offsets iguais a `N*8` = texto vazio.

O texto é encodado no charset do jogo (por idioma) com terminador `0x00`.
Textos idênticos compartilham o mesmo offset (dedup natural da tabela).

## Funções principais

### Leitura (`field_string.go`)

- `FromFieldStringData(bytes, charset, version)` — decodifica o arquivo
  inteiro: valida o primeiro `u16`, conta as entradas e devolve
  `[]*FieldString`.
- `NewFieldString(...)` — monta um `FieldString` a partir dos 8 bytes de
  header + o bloco de texto; resolve os bytes do texto pelo offset.
- `NewEmptyFieldString(charset, version)` — placeholder vazio para
  localizações ausentes.

### Escrita / rebuild (`field_string.go`, `events_strings_binary_writer.go`)

- `RebuildFieldStrings(strings, charset, version)` — reconstrói o bloco de
  texto: os ponteiros são recalculados a partir do comprimento dos textos
  (`contentOffset = N*8`; cada texto novo empurra o seguinte). Deduplicação
  por mapa de offset.
- `EncodeLocalizedStrings(localizedStrings, languageCode, version)` — codec
  compartilhado com os formatos da mesma tabela (ver
  [tabelas-evento.md](tabelas-evento.md)): header de 8 bytes por entrada +
  `RebuildFieldStrings`.
- `writeEventStringsToAllLocalizations(pathPattern, ...)` — grava o binário
  em `mods/<localização>/event/obj_ps3/...` para todos os idiomas.

### Ciclo de vida (`events_binary_file.go`)

- `EventsBinaryFile` — carrega/salva eventos de todas as localizações;
  cada evento tem um arquivo por idioma.
- `ReadLocalizedStringFiles` / `ReadCompleteEventFile` — descoberta e
  leitura dos arquivos de evento na árvore do jogo.
- `SaveToBinary("")` salva tudo; `SaveToBinary(eventID)` salva um evento.

## Relação com os outros formatos

- A estrutura de 8 bytes/entrada e o rebuild são **os mesmos** dos
  binários FFX-2 de batalha/cloudsave/tutorial — ver
  [tabelas-evento.md](tabelas-evento.md).
- Os textos de sistema (`battle/kernel/*.bin`) usam layout DIFERENTE —
  header + chunks + string table — ver
  [textos-objetos.md](textos-objetos.md).
