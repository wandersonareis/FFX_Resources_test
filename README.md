# FINAL FANTASY X and X-2 HD Remaster Resources Editor

<p align="center">
  <img src="docs/imagens/logo_us.png" alt="FINAL FANTASY X | X-2 HD Remaster" width="480">
</p>

O editor de recursos do FINAL FANTASY X / X-2 HD Remaster: extração, tradução
e reimportação dos textos do jogo (eventos, menus, batalha, ajuda, macros),
além de texturas PhyreEngine e o container master `.vbf`.

Construído em Go + [Wails v3](https://v3alpha.wails.io/) (beta) com frontend
Next.js (React 19, App Router) rodando em WebView nativo — sem navegador
embutido, binário único.

## Como funciona

Os binários do jogo são lidos direto da árvore original (`.vbf` incluso) e
as traduções são aplicadas e salvas em `mods/` — a árvore original nunca é
tocada. Cada formato tem seu próprio pacote de leitura/rebuild em
`backend/fileFormats`, com um motor compartilhado de string table
(`backend/core/strings`) que recalcula os ponteiros quando um texto muda de
tamanho.

## Documentação

A documentação dos formatos de arquivo — layout binário de cada um, com
diagramas e as funções principais do pacote Go correspondente — está em
[docs/README.md](docs/README.md). Índice rápido:

- [Motor de string table](docs/motor-strings.md) — cálculo de ponteiros
  compartilhado (`backend/core/strings`)
- [Eventos](docs/formatos/eventos.md) — `event/obj_ps3/*.bin`
- [Textos de sistema](docs/formatos/textos-objetos.md) — `battle/kernel/*.bin`
- [Tabelas de evento FFX-2](docs/formatos/tabelas-evento.md) — btl/cloud/tutorial
- [Painel de ajuda](docs/formatos/ajuda-sps2.md) — `.sps2`
- [Dicionário de macros](docs/formatos/dicionario-macros.md) — `macrodic.dcp`
- [Loc kit](docs/formatos/lockit.md) — `ffx_loc_kit_*.bin`
- [Container `.vbf`](docs/formatos/vbf.md) — leitura-only
- [Texturas `.dds.phyre`](docs/formatos/texturas-phyre.md) — PhyreEngine DX11

## Desenvolvimento

Requisitos: Go 1.26+, [Wails CLI v3](https://v3alpha.wails.io/guides/cli/)
(`wails3`) e [Bun](https://bun.sh) (ou npm — ver `PACKAGE_MANAGER` no
`Taskfile.yml`).

```bash
wails3 dev          # app completo em modo dev (Go + frontend hot reload)
```

O dev server do frontend (Next.js) sobe em `http://localhost:9245` (porta
configurável via `WAILS_VITE_PORT`).

Alternativa via Taskfile (usado pelo CI/IDE):

```bash
task dev
```

## Build

```bash
wails3 build         # binário de produção em build/bin
task build           # equivalente via Taskfile
task package         # build + empacotamento da plataforma
```

O executável permanece em `build/bin` — é de lá que o app resolve
`config/config.json`, `data/`, `extracted/`, `translated/` e `logs/`.

## Testes

```bash
go test ./...
```

Cobrem roundtrip byte a byte de todos os formatos (export → import sem
edição reproduz o binário original), deslocamento de ponteiros ao editar
textos e a guarda contra ponteiros quebrados (MID/OOB) no save.
