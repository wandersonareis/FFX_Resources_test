'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTable } from '@tanstack/react-table';
import { useHotkeys } from '@tanstack/react-hotkeys';
import { useSelector } from '@tanstack/react-store';
import { GitBranch, RotateCcw } from 'lucide-react';
import { dto } from '@/wailsjs/go/models';
import { useEditDraft, rowKey } from '@/lib/ffx/edit-draft';
import { SOURCE_LANG } from '@/lib/ffx/save-all';
import { resolveSegmentLabel } from '@/lib/ffx/display-names';
import { DEDUP_VIEW_KINDS, LINKED_SIDE, linkedRowText } from '@/lib/ffx/hash-ref';
import { GameTextView } from '@/components/game-text-view';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { type EntryView } from './entry-view-store';
import { columnHelper, features } from './table-types';
import { entryNodeId } from '@/lib/ffx/content-tree/model';
import { SHORTCUTS } from '@/lib/ffx/shortcuts';
import { locFromVbfPath } from '@/lib/ffx/tree-data';

/**
 * Célula de uma linha que SÓ EXISTE num dos dois lados (união da tabela):
 * o traço vermelho diz "falta aqui" sem exibir conteúdo do outro lado.
 */
const MISSING_SIDE = 'text-red-600 dark:text-red-400';

/**
 * Tabela de linhas do arquivo selecionado: colunas (Original/Traduzido),
 * navegação por teclado ↑/↓/Enter e foco da 1ª linha pedido pela árvore
 * (ArrowRight). Estado compartilhado (rows, seleção) vem do store da view.
 */
export function EntryTable({ view }: { view: EntryView }) {
  const { store, actions } = view;
  const rows = useSelector(store, (s) => s.rows);
  const activeKind = useSelector(store, (s) => s.activeKind);
  const selectedEntry = useSelector(store, (s) => s.selectedEntry);
  const loading = useSelector(store, (s) => s.loading);
  const refLinks = useSelector(store, (s) => s.refLinks);
  const { store: drafts, snapshot } = useEditDraft();
  const version = view.version;

  // Texto da célula LINKADA (row de ref): estado atual da def — rascunho
  // dela inclusive. O "$hash" do backend nunca é pintado.
  const linkedValueOf = useCallback(
    (row: dto.TextRow): string | undefined => {
      const link = refLinks[rowKey(row)];
      const entry = selectedEntry;
      if (!link) return undefined;
      if (!entry) return link.text ?? '';
      return linkedRowText(entry.id, link, (defId, defKey) =>
        drafts.editTextOf(
          version,
          entry.kind,
          defId,
          defKey,
          SOURCE_LANG,
          entry.vbf?.root ?? ''
        )
      );
    },
    [refLinks, selectedEntry, drafts, version]
  );

  const [focusedRowId, setFocusedRowId] = useState<string | null>(null);
  const rowRefs = useRef(new Map<string, HTMLTableRowElement>());
  /**
   * Wrapper externo da tabela — o `<table>` da shadcn não reencaminha ref. É
   * o ALVO dos hotkeys: o listener só vê eventos que borbulham por aqui (foco
   * dentro da tabela), e o ref nulo (sem linhas servíveis) desliga o alvo.
   */
  const tableRef = useRef<HTMLDivElement>(null);

  // ---- Teclado: tabela — ↑/↓ movem o foco entre linhas; Enter abre o
  // modal do editor da linha focada; ← devolve o foco para a árvore (fecha o
  // ciclo de navegação árvore ↔ tabela).

  /**
   * Linha da tabela sob o evento — qualquer foco dentro dela (a própria `<tr>`
   * ou um botão de célula). `null` = foco fora da tabela.
   */
  const rowElOf = (event: KeyboardEvent): HTMLElement | null => {
    const target = event.target;
    if (!(target instanceof HTMLElement)) return null;
    return target.closest<HTMLElement>('[data-row-key]');
  };

  /**
   * A própria `<tr>` sob o foco — `null` quando o alvo é um botão da célula
   * (Reverter / editar como divergência). Só o Enter usa este guard: sem ele
   * o `preventDefault` mataria o click nativo do botão.
   */
  const rowFocused = (event: KeyboardEvent): HTMLElement | null => {
    const row = rowElOf(event);
    return row && row === event.target ? row : null;
  };

  /** Linhas na MESMA ordem do DOM (== ordem de `rows`, que as renderiza). */
  const rowEls = () =>
    Array.from(
      tableRef.current?.querySelectorAll<HTMLElement>('[data-row-key]') ?? []
    );

  const focusRow = (rowEl: HTMLElement | undefined) => {
    if (!rowEl) return;
    setFocusedRowId(rowEl.dataset.rowKey ?? null);
    rowEl.focus();
  };

  const moveRow = (event: KeyboardEvent, delta: number) => {
    const rowEl = rowElOf(event);
    if (!rowEl) return;
    // Consome a tecla para não rolar a página. Setas não têm uso nativo em
    // `<tr>`/`<button>`, então isso é seguro mesmo com foco na célula.
    event.preventDefault();
    const list = rowEls();
    focusRow(list[list.indexOf(rowEl) + delta]);
  };

  const openFocusedRow = (event: KeyboardEvent) => {
    // Guard estrito: só a linha consome o Enter. Nos botões internos ele
    // segue nativo (aciona o botão) em vez de abrir o editor por cima.
    const rowEl = rowFocused(event);
    if (!rowEl) return;
    event.preventDefault();
    const key = rowEl.dataset.rowKey ?? '';
    const row = rows.find((r) => rowKey(r) === key);
    if (!row) return;
    setFocusedRowId(key);
    // Row de ref dedupada: o editor abre NA DEF (através do link) — nunca em
    // cima do "$hash".
    if (refLinks[key]) actions.openLinked(row);
    else actions.openDialog(row);
  };

  const backToTree = (event: KeyboardEvent) => {
    if (!rowElOf(event)) return;
    event.preventDefault();
    // Volta o foco para o nó selecionado na árvore. Se o nó estiver
    // desmontado (grupo colapsado), cai no primeiro botão visível da árvore.
    // O id de uma folha do .vbf leva caminho absoluto de Windows (barra
    // invertida), então a comparação é pelo atributo — nunca por seletor CSS.
    const tree = document.getElementById('entry-tree');
    const nodeId = selectedEntry ? entryNodeId(selectedEntry) : '';
    const buttons = Array.from(
      tree?.querySelectorAll<HTMLElement>('[data-node-button]') ?? []
    );
    const node =
      buttons.find(
        (button) =>
          button.closest('[data-node-id]')?.getAttribute('data-node-id') ===
          nodeId
      ) ?? buttons[0];
    node?.focus();
  };

  // Callbacks e options são sincronizados a cada render pelo hook, então não
  // há closure velha de `rows`/`actions` (dispensa useCallback).
  useHotkeys(
    [
      {
        hotkey: SHORTCUTS.navigation.down.key,
        callback: (event) => moveRow(event, 1),
        options: { meta: { name: 'Próxima linha', group: 'Tabela' } },
      },
      {
        hotkey: SHORTCUTS.navigation.up.key,
        callback: (event) => moveRow(event, -1),
        options: { meta: { name: 'Linha anterior', group: 'Tabela' } },
      },
      {
        hotkey: 'Enter',
        callback: openFocusedRow,
        options: { meta: { name: 'Abrir o editor da linha', group: 'Tabela' } },
      },
      {
        hotkey: SHORTCUTS.navigation.left.key,
        callback: backToTree,
        options: { meta: { name: 'Voltar para a árvore', group: 'Tabela' } },
      },
    ],
    // `preventDefault`/`stopPropagation` desligados: o cancelamento é manual
    // e só quando o alvo É a linha, e o evento precisa continuar subindo
    // (Delete em document e handlers sintéticos dos ancestres).
    { target: tableRef, preventDefault: false, stopPropagation: false, ignoreInputs: true }
  );

  // lockit: numeração própria por grupo (game/utf8), sem expor o índice
  // técnico do backend.
  const lockitSeq = useMemo(() => {
    const counters = new Map<string, number>();
    const map = new Map<string, number>();
    for (const r of rows) {
      const k = r.name ?? '';
      const n = (counters.get(k) ?? 0) + 1;
      counters.set(k, n);
      map.set(rowKey(r), n);
    }
    return map;
  }, [rows]);

  // Entrada .vbf de idioma não-us: a coluna Original/Traduzido mostra o
  // texto DESTE binário do idioma clicado (não o 'us').
  const textLoc = selectedEntry?.vbf?.path
    ? (locFromVbfPath(selectedEntry.vbf.path) ?? SOURCE_LANG)
    : SOURCE_LANG;

  const columns = useMemo(
    () =>
      columnHelper.columns([
        columnHelper.accessor('index', {
          header: '#',
          cell: (info) => {
            const r = info.row.original;
            if (activeKind === 'lockit') {
              const n = lockitSeq.get(rowKey(r)) ?? '';
              return <span className="text-muted-foreground">{n}</span>;
            }
            return <span className="text-muted-foreground">{info.getValue()}</span>;
          },
        }),
        ...(activeKind === 'objects' || activeKind === 'lockit'
          ? [
              columnHelper.accessor('name', {
                header: activeKind === 'lockit' ? 'Tipo' : 'Nome',
                cell: (info) =>
                  activeKind === 'lockit' ? (
                    <span className="text-muted-foreground">
                      {resolveSegmentLabel(info.getValue())}
                    </span>
                  ) : (
                    <span>{info.getValue()}</span>
                  ),
              }),
            ]
          : []),
        columnHelper.accessor((row) => row.original?.[textLoc], {
          id: 'original',
          header: 'Original',
          cell: (info) => {
            const value = info.getValue();
            // União: linha que só existe na tradução fica sem Original e é
            // marcada — o traço vermelho diz "falta aqui" sem mostrar texto
            // do outro lado. Sem contraparte o valor é undefined → "—".
            // NUNCA cai para row.text: seria exibir a tradução como original.
            const missing = info.row.original.missingInOriginal === true;
            return (
              <GameTextView
                text={value ?? ''}
                fallback={value === undefined ? '—' : ''}
                className={missing ? MISSING_SIDE : undefined}
              />
            );
          },
        }),
        columnHelper.accessor(
          (row) => {
            // Ref dedupada: a célula pinta o TEXTO DA DEF (estado atual,
            // rascunho incluído) — o "$hash" entregue pelo backend é só o
            // ponteiro de collapse, nunca conteúdo exibível.
            const entry = selectedEntry;
            // Rascunho por FONTE: data/ ("") e o .vbf (caminho do container)
            // têm namespaces separados. Ler sem a fonte esconderia a edição
            // feita pela ref (na def de outra entrada — inclusive de .vbf).
            const edited = entry
              ? drafts.editOf(
                  version,
                  entry.kind,
                  entry.id,
                  row,
                  SOURCE_LANG,
                  entry.vbf?.root ?? ''
                )
              : undefined;
            const divergent = entry
              ? drafts.isDivergent(
                  version,
                  entry.kind,
                  entry.id,
                  row,
                  entry.vbf?.root ?? ''
                )
              : false;
            // Enquanto a escolha de divergência está no rascunho, a edição
            // local vence o link e a célula deixa de parecer azul.
            if (divergent && edited !== undefined) return edited;
            const linked = linkedValueOf(row);
            if (linked !== undefined) return linked;
            return edited ?? row.text?.[textLoc] ?? '';
          },
          {
            id: 'translated',
            header: 'Traduzido',
            cell: (info) => {
              const row = info.row.original;
              // União: linha que só existe no original tem o texto vazio e a
              // marcação — nada a traduzir, então o diálogo não abre.
              const missing = row.missingInTranslated === true;
              const entry = selectedEntry;
              // Somente leitura = .vbf de kind fora da lista de kinds
              // editáveis (nenhum texto hoje: events, tabelas eventtable,
              // help, macro, objects e lockit editam pelo .vbf, gravando
              // sempre em mods/). O rascunho é da fonte do container.
              const readOnly = entry?.vbf != null;
              const edited =
                !missing && entry
                  ? drafts.editOf(
                      version,
                      entry.kind,
                      entry.id,
                      row,
                      SOURCE_LANG,
                      entry.vbf?.root ?? ''
                    ) !== undefined
                  : false;
              const divergentDraft =
                !missing && entry
                  ? drafts.isDivergent(
                      version,
                      entry.kind,
                      entry.id,
                      row,
                      entry.vbf?.root ?? ''
                    )
                  : false;
              if (missing) {
                return (
                  <div className="text-cell translated-cell" title="Linha inexistente no arquivo traduzido">
                    <span className={MISSING_SIDE}>—</span>
                  </div>
                );
              }
              // Ref dedupada (link): texto da def em cor própria, com o
              // original dela no tooltip. O clique abre o editor NA DEF —
              // a tradução é registrada no arquivo da def, nunca aqui.
              const link = divergentDraft ? undefined : refLinks[rowKey(row)];
              if (link) {
                const defText = linkedValueOf(row) ?? link.text ?? '';
                // Fallback: original ausente (def sem pristine em data/)
                // degrada para o texto atual da def em vez de truncar o
                // tooltip ("Repetição de: " sem texto).
                const original = link.original || link.text || '';
                const defTranslated = defText !== '' && defText !== original;
                return (
                  <div className="group flex items-center gap-1">
                    <div
                      className={`text-cell translated-cell min-w-0 flex-1${defTranslated ? ' edited' : ''}`}
                      title={
                        defTranslated
                          ? `Repetição de: ${original}`
                          : 'Repetição (dedup): o texto vive na def de outro ponto'
                      }
                      onClick={
                        readOnly ? undefined : () => actions.openLinked(row)
                      }
                    >
                      <GameTextView
                        text={defText}
                        fallback="—"
                        className={LINKED_SIDE}
                      />
                    </div>
                    {!readOnly ? (
                      <button
                        type="button"
                        aria-label="Editar como divergência nesta cópia"
                        title="Editar como divergência nesta cópia"
                        className="shrink-0 rounded p-1 text-muted-foreground opacity-60 transition-opacity hover:bg-muted hover:text-foreground hover:opacity-100 focus:opacity-100 group-hover:opacity-100"
                        onClick={(event) => {
                          event.stopPropagation();
                          actions.openLinked(row, true);
                        }}
                      >
                        <GitBranch size={15} aria-hidden />
                      </button>
                    ) : null}
                  </div>
                );
              }
              const displayed = String(info.getValue() ?? '');
              const original = row.original?.[SOURCE_LANG];
              const canRevert =
                !readOnly &&
                textLoc === SOURCE_LANG &&
                original !== undefined &&
                original !== '' &&
                displayed !== original;
              return (
                <div className="group flex items-center gap-1">
                  <div
                    className={`text-cell translated-cell min-w-0 flex-1${edited || canRevert ? ' edited' : ''}`}
                    onClick={
                      readOnly ? undefined : () => actions.openDialog(row)
                    }
                  >
                    <GameTextView text={displayed} />
                  </div>
                  {canRevert ? (
                    <button
                      type="button"
                      aria-label="Reverter esta linha para o original"
                      title="Reverter esta linha para o original (se for cópia, volta a convergir com a definição)"
                      className="shrink-0 rounded p-1 text-muted-foreground opacity-60 transition-opacity hover:bg-muted hover:text-foreground hover:opacity-100 focus:opacity-100 group-hover:opacity-100"
                      onClick={(event) => {
                        event.stopPropagation();
                        actions.revertToOriginal(row);
                      }}
                    >
                      <RotateCcw size={15} aria-hidden />
                    </button>
                  ) : null}
                </div>
              );
            },
          }
        ),
      ]),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [activeKind, selectedEntry, version, snapshot.revision, drafts, actions, lockitSeq, refLinks, linkedValueOf]
  );

  const table = useTable({
    features,
    columns,
    data: rows,
    // A união da tabela pode ter duas rows com o mesmo índice (uma em cada
    // lado) — o id da row precisa do nome junto para não colidir no React.
    getRowId: (row) => rowKey(row),
  });

  const pendingTableFocus = useSelector(store, (s) => s.pendingTableFocus);

  // ArrowRight na folha: quando o arquivo termina de carregar, foca a
  // primeira linha (apenas focus() no DOM.
  useEffect(() => {
    if (!pendingTableFocus || !selectedEntry || rows.length === 0) return;
    actions.consumeTableFocus();
    rowRefs.current.get(rowKey(rows[0]))?.focus();
  }, [pendingTableFocus, selectedEntry, rows, actions]);

  // Arquivo sem NENHUMA linha servível: a tabela ficaria com o corpo vazio,
  // então o aviso explica. Note que as refs NÃO entram aqui — elas ficam
  // VISÍVEIS como link (nada é oculto: nem arquivo, nem linha repetida, em
  // data/ ou no .vbf).
  if (
    DEDUP_VIEW_KINDS.has(activeKind) &&
    selectedEntry &&
    !loading &&
    rows.length === 0
  ) {
    return (
      <div className="mt-2 rounded-md border border-dashed px-3 py-6 text-center text-sm text-muted-foreground">
        Nenhuma linha traduzível neste arquivo (todas em branco no original).
        As repetições de textos definidos em outro arquivo saem como link para
        a def — nada é ocultado.
      </div>
    );
  }

  return (
    <div className="mt-2" ref={tableRef}>
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <TableHead
                  key={header.id}
                  className={
                    header.column.id === 'index' || header.column.id === 'seq'
                      ? 'w-14'
                      : undefined
                  }
                >
                  {header.isPlaceholder ? null : <table.FlexRender header={header} />}
                </TableHead>
              ))}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {table.getRowModel().rows.map((tRow) => {
            const r = tRow.original;
            const key = rowKey(r);
            return (
              <TableRow
                key={tRow.id}
                ref={(el) => {
                  if (el) rowRefs.current.set(key, el);
                  else rowRefs.current.delete(key);
                }}
                tabIndex={0}
                data-row-key={key}
                onFocus={() => setFocusedRowId(key)}
                className={
                  focusedRowId === key
                    ? 'bg-sky-100 outline outline-sky-400'
                    : 'outline-none focus-visible:bg-muted/40'
                }
              >
                {tRow.getAllCells().map((cell) => (
                  <TableCell
                    key={cell.id}
                    className={cell.column.id === 'original' ? 'text-cell' : undefined}
                  >
                    <table.FlexRender cell={cell} />
                  </TableCell>
                ))}
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}
