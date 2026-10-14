# Loc kit (`ffx_loc_kit_*.bin`)

Pacote: `backend/fileFormats/lockit`

Arquivos: `ffx_data/gamedata/ps3data/lockit/ffx_loc_kit_ps3_<lang>.bin` —
uma variante por idioma.

## Layout do arquivo

É um conjunto de registros de texto separados por **CRLF**, que mistura
**duas codificações no MESMO arquivo**:

- `game`: bytes no charset customizado do jogo (`backend/core/encoding`),
  cuja tabela depende do idioma (us/default/jp/kr/ch);
- `utf8`: texto UTF-8 puro (launcher/PC).

```
+--------------------------------+
| registro 1 (bytes) \r\n        |  kind fixo por posição:
| registro 2 (bytes) \r\n        |  a estrutura (qual registro é game
| ...                            |  e qual é utf8) é IDÊNTICA entre
| registro N (bytes) \r\n?       |  idiomas — detectada por votação
+--------------------------------+
```

- A estrutura de kinds é fixa por versão: o mesmo registro é `game` em
  todos os idiomas (detectado comparando os arquivos de todos os idiomas).
- `terminated` = se o arquivo termina com CRLF (preservado no rebuild).

## Funções principais

### Layout / paths

- `Layout` (`PatternPath()`, `Key()`, `ID()`, `RelPath(lang)`...) —
  resolve os caminhos do arquivo por idioma.
- `Kind` (`String()`, `KindFromName()`) — rótulos `game`/`utf8` do DTO.

### Leitura

- `Load(l)` / `LoadFrom(l, src)` — lê TODOS os idiomas de uma vez e devolve
  o `LockitFile` lógico único: registros com texto por idioma.
- `detectKinds(rawByLang, n)` — votação entre idiomas para decidir o kind
  de cada registro.
- `decodeRecord(raw, kind, lang, version)` — decodifica um registro (charset
  do jogo ou UTF-8).

### Escrita

- `Record.SetText(lang, s)` — edita o texto de um idioma (bytes originais
  preservados por idioma: o encode do jogo tem representações
  alternativas).
- `f.Bytes()` / `f.Save()` / `f.SaveLanguages(langs...)` — re-serializa e
  grava em `mods/` (um arquivo por idioma editado).
- `encodeRecord(text, kind, lang, version)` — recodifica um registro.
- `f.IndexesByKind()` — índices dos registros game e utf8.

## Detalhe importante

Como cada idioma é um arquivo separado mas a estrutura de kinds é
compartilhada, a extração agrupa tudo em **um único arquivo lógico** no
DTO, com todos os idiomas por row — e o save só reescreve os idiomas que
mudaram.
