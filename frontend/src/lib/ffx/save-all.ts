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
