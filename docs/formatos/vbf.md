# Container Virtuos Big File (`.vbf`)

Pacote: `backend/fileFormats/vbf`

Arquivos: `FFX_Data.vbf` e `FFX2_Data.vbf` — o container master dos
remasters (Virtuos). **SOMENTE LEITURA**: o pacote nunca abre nada em
escrita e nunca materializa a árvore inteira — só o cabeçalho (índice) e
os blocos do arquivo pedido, sob demanda.

## Layout do arquivo

```
u32 magic "SRYK" (0x4B595253)
u32 headerLen                    -- bytes do cabeçalho, do offset 0 ao fim
                                   da lista de blocos (coberto pelo MD5)
u64 numFiles
numFiles × [16] MD5(path)        -- hash do caminho (conferência/lookup)
numFiles × 32:  startBlock u32, 0 u32, bytes u64, offset u64,
                nameOffset u64
u32 nameTableLen (inclui os próprios 4 bytes) + nameTable
blocks × u16                     -- tamanho comprimido de cada bloco
                                   (0 = 64 KiB crú)
⟨payload dos arquivos⟩
[16] MD5(header[0:headerLen])    -- últimos 16 bytes do arquivo
```

O payload de um arquivo é uma sequência de blocos de 65536 bytes de
saída. Cada bloco é **CRU** quando o `u16` vale 0 (mapeado para 65536) e
**deflate RAW** caso contrário (os 2 bytes de cabeçalho zlib são pulados —
é assim que o leitor de referência descompacta).

> O layout foi reconstruído empiricamente: formato de autoria
> desconhecida, decompilado de uma ferramenta de origem incerta e do
> leitor embutido no launcher — conferido contra os `.vbf` reais
> (assinatura, índice, blocos e MD5 do cabeçalho).

## Funções principais

- `Open(filePath)` — valida magic e MD5 do header, carrega o índice;
  devolve `*Archive`.
- `a.Entry(name)` / `a.Has(name)` / `a.IsDir(dir)` — lookup por nome
  normalizado.
- `a.List(dir)` / `a.Paths()` / `a.FilesUnder(dir)` — navegação da árvore
  virtual.
- `a.Read(name)` / `a.ReadEntry(e)` — lê um arquivo inteiro (descomprime
  os blocos sob demanda).
- `a.ExtractEntry(e, destPath)` — extrai para disco (com limite de
  tamanho; sem deixar `.tmp`).
- `a.Close()`, `a.Path()`, `a.Len()`, `a.Size()` — metadados.

Erros esperados: `ErrNotVBF`, `ErrHeaderMD5Invalido`, `ErrNotFound`,
`ErrTooLarge`.

## Uso na app

O `.vbf` é a fonte ORIGINAL dos binários do jogo: os pacotes de formato
resolvem os arquivos dentro do container via `common.FileAccessor`/
`FileSource` quando o usuário aponta para o jogo instalado, e os edits são
salvos em `mods/` fora do container.
