# Motor de string table (`backend/core/strings`)

O cálculo de ponteiros de string table é **um motor só**, compartilhado
por todos os formatos de objetos textuais ([objetos de
sistema](formatos/textos-objetos.md), [w_name](formatos/textos-objetos.md),
[ajuda](formatos/ajuda-sps2.md)). A lógica é idêntica entre eles: **um
bloco de texto e o comprimento do texto geram o valor dos ponteiros**.

## Contrato `TableRef`

Qualquer objeto `datastore.IGlobalKeyedString` participa do motor:
offset (`u16`), key (`u8`), bytes armazenados, charset e status do
ponteiro. O chunkmap e o helpfile criam `KeyedString`s que implementam o
contrato.

## Status de ponteiro (`PointerStatus`)

Um ponteiro `u16 offset` contra a string table pode estar em 5 estados:

| Estado | Significado |
|---|---|
| `PointerOK` | aponta para texto válido (terminador 0x00 encontrado) |
| `PointerEmpty` | aponta para a cauda da tabela = sem texto |
| `PointerMid` | aponta para o MEIO de um texto (quebrado) |
| `PointerOOB` | offset fora da tabela (quebrado) |
| `PointerCreated` | campo criado na edição (não existia no original) |

- `PointerStatusAt(table, offset)` — classifica um offset contra a tabela.
- `PointerStatus.IsBroken()` — `Mid`/`OOB` (o jogo leria texto inválido).
- `PointerStatusAt` é a base dos testes de medição de `event` e `macrodic`
  e do guard do chunkmap.

## Rebuild (`RebuildKeyedStrings`)

```
RebuildKeyedStrings(list, charset, version, base) []byte
```

Reconstrói a string table a partir da lista editada de keyed strings:

1. **Splice-on-base**: quando possível, os textos editados são inseridos
   NO bloco de texto original (`base`), preservando os bytes ao redor — os
   offsets dos textos não editados não mudam.
2. **Cauda**: textos vazios apontam para a cauda da tabela (após o último
   terminador).
3. **Dedup**: textos idênticos compartilham offset (mapa texto→offset).
4. Fallback: se o splice não é utilizável, a tabela é reconstruída do
   zero (`rebuildStringTableFromScratch`) e todos os offsets recalculados.

Ponteiros quebrados (MID/OOB) **recebem** o deslocamento global sem
causá-lo, e nunca viram ponteiros válidos no save.

- `UsableStringTableBase(list, candidate)` — decide se o `base` original
  ainda serve como base de splice.
- `safeTailPadding`, `blockStarts`, `originalBlockLen` — utilidades
  internas do splice/cauda.

## `KeyedString`

- `NewKeyedString(charset, segment, data, version)` — cria a partir de um
  segmento e da tabela; classifica o status na criação.
- `SetString(str, charset)` — edita o texto (recodifica no charset).
- `GetString`/`StoredBytes`/`Edited`/`PointerStatus` — acesso ao estado.

## Garantias (testes)

- **Identidade**: sem edição, o rebuild reproduz a string table byte a
  byte (roundtrips de todos os formatos).
- **+1 caractere em um texto**: todos os ponteiros posteriores andam +1;
  os anteriores ficam parados
  (`TestEditingOneCharacterMovesFollowingTextPointersByOne`).
- **Deslocamento global**: editar um campo move todos os segmentos
  posteriores pelo mesmo delta, e os quebrados continuam quebrados
  (`checkSaveKeepsSegments` — arquivos reais de todas as versões).
