'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTable } from '@tanstack/react-table';
import { useSelector } from '@tanstack/react-store';
import { dto } from '@/wailsjs/go/models';
import { useEditDraft } from '@/lib/ffx/edit-draft';
import { SOURCE_LANG } from '@/lib/ffx/save-all';
import { GameTextView } from '@/components/game-text-view';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import type { EntryView } from './entry-view-store';
import { columnHelper, features } from './types';

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
  const { store: drafts, snapshot } = useEditDraft();
  const version = view.version;

  // Navegação por teclado: linha focada na tabela (refs p/ mover o foco).
  const [focusedRowId, setFocusedRowId] = useState<string | null>(null);
  const rowRefs = useRef(new Map<string, HTMLTableRowElement>());

  // ---- Teclado: tabela — ↑/↓ movem o foco entre linhas; Enter abre o
  // modal do editor da linha focada.
  const onRowKeyDown = useCallback(
    (
      event: React.KeyboardEvent<HTMLTableRowElement>,
      row: dto.TextRow,
      rowKey: string
    ) => {
      const idx = rows.findIndex((r) => r.index === row.index);
      if (event.key === 'ArrowDown') {
        event.preventDefault();
        const next = rows[idx + 1];
        if (next) {
          const key = String(next.index);
          setFocusedRowId(key);
          rowRefs.current.get(key)?.focus();
        }
      } else if (event.key === 'ArrowUp') {
        event.preventDefault();
        const prev = rows[idx - 1];
        if (prev) {
          const key = String(prev.index);
          setFocusedRowId(key);
          rowRefs.current.get(key)?.focus();
        }
      } else if (event.key === 'Enter') {
        event.preventDefault();
        setFocusedRowId(rowKey);
        actions.openDialog(row);
      }
    },
    [rows, actions]
  );

  const columns = useMemo(
    () =>
      columnHelper.columns([
        columnHelper.accessor('index', {
          header: '#',
          cell: (info) => <span className="text-muted-foreground">{info.getValue()}</span>,
        }),
        ...(activeKind === 'objects'
          ? [columnHelper.accessor('name', { header: 'Nome' })]
          : []),
        columnHelper.accessor((row) => row.text?.[SOURCE_LANG] ?? '', {
          id: 'original',
          header: 'Original',
          cell: (info) => <GameTextView text={info.getValue()} />,
        }),
        columnHelper.accessor(
          (row) => {
            const entry = selectedEntry;
            const edited = entry
              ? drafts.editOf(version, entry.kind, entry.id, row, SOURCE_LANG)
              : undefined;
            return edited ?? row.text?.[SOURCE_LANG] ?? '';
          },
          {
            id: 'translated',
            header: 'Traduzido',
            cell: (info) => {
              const row = info.row.original;
              const entry = selectedEntry;
              const edited = entry
                ? drafts.editOf(version, entry.kind, entry.id, row, SOURCE_LANG) !==
                  undefined
                : false;
              return (
                <div
                  className={`text-cell translated-cell${edited ? ' edited' : ''}`}
                  onClick={() => actions.openDialog(row)}
                >
                  <GameTextView text={info.getValue()} />
                </div>
              );
            },
          }
        ),
      ]),
    // snapshot.revision: recompute as colunas quando o rascunho muda
    // (editOf lê do store mutável, fora do escopo do memo).
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [activeKind, selectedEntry, version, snapshot.revision, drafts, actions]
  );

  const table = useTable({
    features,
    columns,
    data: rows,
  });

  const pendingTableFocus = useSelector(store, (s) => s.pendingTableFocus);

  // ArrowRight na folha: quando o arquivo termina de carregar, foca a
  // primeira linha (apenas focus() no DOM — o destaque vem do onFocus,
  // evitando setState síncrono dentro do effect).
  useEffect(() => {
    if (!pendingTableFocus || !selectedEntry || rows.length === 0) return;
    actions.consumeTableFocus();
    rowRefs.current.get(String(rows[0].index))?.focus();
  }, [pendingTableFocus, selectedEntry, rows, actions]);

  return (
    <div className="mt-2">
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <TableHead
                  key={header.id}
                  className={header.column.id === 'index' ? 'w-14' : undefined}
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
            const rowKey = String(r.index);
            return (
              <TableRow
                key={tRow.id}
                ref={(el) => {
                  if (el) rowRefs.current.set(rowKey, el);
                  else rowRefs.current.delete(rowKey);
                }}
                tabIndex={0}
                onKeyDown={(event) => onRowKeyDown(event, r, rowKey)}
                onFocus={() => setFocusedRowId(rowKey)}
                className={
                  focusedRowId === rowKey
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
