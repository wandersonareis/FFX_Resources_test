'use client';

import { useEffect } from 'react';
import { useTable } from '@tanstack/react-table';
import { useSelector } from '@tanstack/react-store';
import { rowKey } from '@/lib/ffx/edit-draft';
import { DEDUP_VIEW_KINDS } from '@/lib/ffx/hash-ref';
import { Table } from '@/components/ui/table';
import { type EntryView } from '../entry-view-store';
import { features } from './table-types';
import { useEntryTableHotkeys } from './use-entry-table-hotkeys';
import { useEntryTableColumns } from './use-entry-table-columns';
import { EntryTableHeader } from './entry-table-header';
import { EntryTableBody } from './entry-table-body';
import { EntryTableEmpty } from './entry-table-empty';

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
  const version = view.version;

  const { tableRef, rowRefs, focusedRowId, setFocusedRowId } =
    useEntryTableHotkeys({ rows, refLinks, selectedEntry, actions });

  const columns = useEntryTableColumns({
    rows,
    activeKind,
    selectedEntry,
    refLinks,
    version,
    actions,
  });

  const table = useTable({
    features,
    columns,
    data: rows,
    // A união da tabela pode ter duas rows com o mesmo índice (uma em cada
    // lado) — o id da row precisa do nome junto para não colidir no React.
    getRowId: (row) => rowKey(row),
  });

  const pendingTableFocus = useSelector(store, (s) => s.pendingTableFocus);
  const pendingTableRowFocus = useSelector(store, (s) => s.pendingTableRowFocus);

  // Um resultado de busca carrega o locator da row junto com o arquivo: em
  // vez de começar no topo, a tabela foca e centraliza a ocorrência encontrada.
  useEffect(() => {
    if (!pendingTableRowFocus || loading || !selectedEntry || rows.length === 0) return;
    const key = rowKey({
      index: pendingTableRowFocus.index,
      name: pendingTableRowFocus.name,
    });
    const row = rowRefs.current.get(key) ?? rowRefs.current.get(rowKey(rows[0]));
    actions.consumeTableRowFocus();
    if (!row) return;
    setFocusedRowId(row.dataset.rowKey ?? key);
    row.focus();
    row.scrollIntoView({ block: 'center' });
  }, [pendingTableRowFocus, loading, selectedEntry, rows, rowRefs, actions]);

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
    return <EntryTableEmpty />;
  }

  return (
    <div className="mt-2" ref={tableRef}>
      <Table>
        <EntryTableHeader table={table} />
        <EntryTableBody
          table={table}
          rowRefs={rowRefs}
          focusedRowId={focusedRowId}
          onFocusRow={setFocusedRowId}
        />
      </Table>
    </div>
  );
}
