# Painel de ajuda (`.sps2`)

Pacote: `backend/fileFormats/helpfile`

Arquivos: `help/now_help/now_help.sps2` e `now_help_page.sps2` (FFX), e as
variantes FFX-2 (`mon_boku`, `s_monitor`, `dvdcopy`, ...).

## Layout do arquivo

Header fixo de 0x24 bytes (campos `u32` little-endian), seguido de
**tabela de ponteiros u32** e **segmentos de texto**:

```
0x00 magic (1)
0x04 pageCount (0 nos arquivos simples; N nos *_page)
0x08 textEnd   -- fim do bloco de texto (recalculado no rebuild)
0x0C ptrBase   -- 0x24 quando há sub-header de páginas; 0 caso contrário
0x10 ptrTableOff -- início da tabela de ponteiros (0x24 ou 0x114)
0x14 footerPtr -- início da árvore/footer de páginas; 0 se ausente
0x18..0x24 timestamps (12 bytes preservados verbatim)

[ptrBase..ptrTableOff]    sub-header de páginas (preservado verbatim)
[ptrTableOff..+4N]        tabela de N ponteiros u32
                          (N = (ptr[0] - ptrTableOff) / 4)
[ptr[0]..textEnd]         segmentos: texto encodado + 0x00
                          (+ bloco de dados opcional no segmento)
[textEnd..footerPtr]      registros de páginas {u32 offset; u16 size;
                          u16 FFFF} + dados das páginas (offsets
                          deslocados no rebuild)
[footerPtr..EOF]          árvore/footer (preservado verbatim)
```

Cada segmento contém o terminador `0x00`: o texto vai até o `0x00`; o que
sobra até o próximo ponteiro é o bloco de dados do segmento, preservado
byte a byte no rebuild.

## Funções principais

### Leitura

- `ReadHelpBinary(data, ...)` — decodifica o `.sps2` em segmentos
  (`HelpSegment`): texto + dados originais.
- `ReadHelpFile(version, localization)` / `ReadHelpPanelFrom(...)` — lê um
  painel específico da árvore do jogo.
- `HelpEntryNames()`, `IsHelpEntry(name)`, `HelpRelPath(...)`,
  `HelpPathForVersion(...)`, `HelpModsPathForVersion(...)` — resolução de
  caminhos e nomes válidos de painéis.

### Escrita / rebuild

- `RebuildHelpBinary(...)` — re-serializa o arquivo: recalcula `textEnd`,
  desloca os ponteiros da tabela (u32) conforme o novo tamanho dos textos,
  desloca os offsets dos registros de página
  (`shiftPageRecordOffsets`) e preserva sub-header/footer/timestamps.
- `HelpStringsBinaryFile` — ciclo de vida (load/save por localização);
  `HelpKeyedStringFile` — objetos `IGlobalKeyedString` para o
  [motor de strings](../motor-strings.md).

### Store

- `GetHelp` / `SetHelp` / `HasHelp` / `EnsureHelpLoaded` /
  `StampAllPanels` / `EnsureHelpFresh` — cache e freshness por painel.
