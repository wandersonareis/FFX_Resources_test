import { ApplyTextCollection } from '@/wailsjs/go/main/App';
import { toast } from 'sonner';
import { editDraft } from './edit-draft';
import { sendErrorNotification } from './error-handler';
import type { GameVersionId } from './game-version';
import { EntryKind } from './display-names';

export const SOURCE_LANG = 'us';

export async function saveAllDrafts(
  isSaving: () => boolean,
  setSaving: (value: boolean) => void
): Promise<void> {
  if (!editDraft.getSnapshot().hasDirty || isSaving()) return;
  setSaving(true);
  try {
    // Gate de tags protegidas: restaura o que sumiu e aborta se ainda houver
    // valor perdido ou `{…` aberto — o texto não vai ao binário quebrado.
    const report = editDraft.restoreProtectedTags();
    if (report.restored > 0) {
      toast(
        `Tags protegidas restauradas em ${report.restored} célula(s).`,
        { duration: 4000 }
      );
    }
    if (report.unresolved > 0 || report.unclosed > 0) {
      const parts: string[] = [];
      if (report.unresolved > 0) {
        parts.push(
          `${report.unresolved} célula(s) sem tag protegida obrigatória`
        );
      }
      if (report.unclosed > 0) {
        parts.push(`${report.unclosed} célula(s) com tag aberta ({…)`);
      }
      toast.error(
        `Salvamento bloqueado: ${parts.join(' e ')}. Abra a linha e use Restaurar/Corrigir.`,
        { duration: 8000 }
      );
      return;
    }
    for (const [version, byKind] of editDraft.dirtyBatches()) {
      for (const kind of byKind.keys()) {
        const collection = editDraft.buildCollection(
          version as GameVersionId,
          kind as EntryKind
        );
        if (Object.keys(collection).length === 0) continue;
        await ApplyTextCollection(
          kind,
          version as Parameters<typeof ApplyTextCollection>[1],
          collection as Parameters<typeof ApplyTextCollection>[2]
        );
      }
    }
    editDraft.markSaved();
    toast('Alterações salvas no jogo.', { duration: 3000 });
  } catch (error) {
    sendErrorNotification(error);
  } finally {
    setSaving(false);
  }
}
