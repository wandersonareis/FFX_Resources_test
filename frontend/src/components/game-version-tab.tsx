'use client';

import { useEffect, useMemo, useState } from 'react';
import { Save } from 'lucide-react';
import { useSelector } from '@tanstack/react-store';
import { SetUnsavedEdits } from '@/wailsjs/go/main/App';
import { KIND_LABELS } from '@/lib/ffx/display-names';
import type { GameVersionId } from '@/lib/ffx/game-version';
import { useEditDraft } from '@/lib/ffx/edit-draft';
import { useWailsEvent } from '@/lib/ffx/use-wails-event';
import { saveAllDrafts } from '@/lib/ffx/save-all';
import { Button } from '@/components/ui/button';
import {
  createEntryView,
  type EntryView,
} from '@/components/entry-view/entry-view-store';
import { ContentTree } from '@/components/entry-view/content-tree';
import { EntryTable } from '@/components/entry-view/entry-table';
import { TranslationDialog } from '@/components/entry-view/translation-dialog';
import { PreloadVersions } from '@/wailsjs/go/main/App';
import { GAME_VERSIONS } from '@/lib/ffx/game-version';

/**
 * Aba de uma versão do jogo: só orquestra. O estado compartilhado entre
 * árvore, tabela e diálogo vive no store da view (createEntryView) — um por
 * aba — e cada filho se assina com useSelector.
 */
export function GameVersionTab({ version }: { version: GameVersionId }) {
  const view: EntryView = useMemo(() => createEntryView(version), [version]);
  const { store: drafts, snapshot } = useEditDraft();
  const activeKind = useSelector(view.store, (s) => s.activeKind);
  const selectedEntry = useSelector(view.store, (s) => s.selectedEntry);
  const loading = useSelector(view.store, (s) => s.loading);
  const [saving, setSaving] = useState(false);

  const hasDirty = snapshot.hasDirty;

  useEffect(() => {
    void SetUnsavedEdits(hasDirty);
  }, [hasDirty]);

  useEffect(() => {
    // Sincroniza com o backend ao montar/trocar de versão. Ao concluir a
    // carga da aba VISUALIZADA, pré-carrega as demais em background —
    // trocar de aba depois cai no caminho rápido (cache do backend).
    void view.actions.reload().then(() => {
      void PreloadVersions(
        GAME_VERSIONS.filter((t) => t.id !== version).map((t) => t.id)
      );
    });
  }, [view, version]);

  useEffect(
    () => drafts.onSaved(() => void view.actions.reload()),
    [drafts, view]
  );

  // Após importar, o backend emite ImportDone → recarrega árvore e tabela.
  useWailsEvent('ImportDone', () => void view.actions.reload());

  const onSaveAll = () => {
    void saveAllDrafts(() => saving, setSaving);
  };

  return (
    <div className="flex h-full min-h-0 border-t">
      <ContentTree view={view} />

      <main className="flex-1 min-w-0 p-3 px-4 overflow-auto">
        <div className="flex items-center justify-between gap-4">
          <h3 className="text-lg font-semibold">
            {KIND_LABELS[activeKind]}
            {selectedEntry ? (
              <span className="font-normal opacity-70"> · {selectedEntry.label}</span>
            ) : null}
          </h3>
          <Button disabled={!hasDirty || saving} onClick={() => void onSaveAll()}>
            <Save size={18} />
            {saving ? 'Salvando…' : 'Salvar'}
          </Button>
        </div>

        {selectedEntry ? (
          <EntryTable view={view} />
        ) : loading ? (
          <p className="mt-8 opacity-70">Carregando…</p>
        ) : (
          <p className="mt-8 opacity-70">Selecione um arquivo no sidebar.</p>
        )}
      </main>

      <TranslationDialog view={view} />
    </div>
  );
}
