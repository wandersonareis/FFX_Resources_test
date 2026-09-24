'use client';

import { useEffect, useState } from 'react';
import { useSelector } from '@tanstack/react-store';
import { ListLanguages } from '@/wailsjs/go/main/App';
import { SOURCE_LANG } from '@/lib/ffx/save-all';
import { TranslationCellDialog } from '@/components/dialogs/translation-cell-dialog';
import type { EntryView } from './entry-view-store';

/**
 * Fio do diálogo de tradução com o store da view: busca os idiomas de
 * consulta uma única vez e liga navegação/fechamento às actions (que gravam
 * o rascunho via editDraft e avançam a translationRow).
 */
export function TranslationDialog({ view }: { view: EntryView }) {
  const { store, actions } = view;
  const open = useSelector(store, (s) => s.dialogOpen);
  const row = useSelector(store, (s) => s.translationRow);
  const rows = useSelector(store, (s) => s.rows);
  const [languages, setLanguages] = useState<Array<{ code: string; name: string }>>([]);

  useEffect(() => {
    ListLanguages()
      .then((v) => setLanguages(v ?? [{ code: SOURCE_LANG, name: 'English' }]))
      .catch(() => setLanguages([{ code: SOURCE_LANG, name: 'English' }]));
  }, []);

  return (
    <TranslationCellDialog
      open={open}
      row={row}
      languages={languages}
      hasPrevious={rows.findIndex((r) => r.index === row?.index) > 0}
      hasNext={
        rows.findIndex((r) => r.index === row?.index) < rows.length - 1
      }
      onNavigate={(direction, value) => actions.navigateRow(direction, value)}
      onClosed={(value) => actions.commitRow(value)}
    />
  );
}
