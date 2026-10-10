'use client';

import { useState } from 'react';
import { loggedToast as toast } from '@/lib/ffx/toast-logged';
import type { dto } from '@/wailsjs/go/models';
import type { EntryRow } from '@/lib/ffx/tree-data';
import {
  importImage,
  importImageGroup,
  refreshImageDuplicates,
  saveImage,
  selectImageFile,
  selectImageSavePath,
} from '@/lib/ffx/tree-data';
import {
  extractVbfImagesSelection,
  importVbfImage,
  invalidateVbfImage,
  saveVbfImage,
} from '@/lib/ffx/vbf';
import { parseError } from '@/lib/ffx/error-handler';
import type { EntryView } from '@/components/entry-view/entry-view-store';

export type ImageBusy = 'extract' | 'import' | 'save' | 'refresh' | null;

/**
 * Fluxos imperativos do painel: salvar, importar (com passo de alcance
 * quando há cópias), reanalisar duplicatas e abrir ação de imagem.
 *
 * Todo o "chama binding → grava → toast → recarrega" vive aqui; o JSX fica
 * só com o que pintar. `busy` trava os botões durante a gravação e
 * `pendingImport` é o .dds já escolhido aguardando a escolha de alcance.
 */
export function useImageActions({
  view,
  entry,
  duplicates,
}: {
  view: EntryView;
  /** null = nenhuma textura selecionada (o painel não renderiza nada). */
  entry: EntryRow | null;
  duplicates: dto.ImageDuplicate[];
}) {
  const { version, actions } = view;
  const [busy, setBusy] = useState<ImageBusy>(null);
  /**
   * .dds já escolhido aguardando o usuário decidir o alcance do import
   * (null = sem diálogo aberto). Só aparece quando a imagem tem cópias.
   */
  const [pendingImport, setPendingImport] = useState<string | null>(null);

  /**
   * Textura aberta pelo navegador de .vbf: o container é somente leitura;
   * extrair usa o caminho interno e importar/salvar grava fora dele. Delete e
   * replicate de grupos continuam restritos à árvore data/.
   */
  const readOnly = Boolean(entry?.vbf);

  const refuseVbf = (): boolean => {
    if (!readOnly) return false;
    toast.message(
      'Exclusão e replicação estão disponíveis apenas para imagens da árvore de data/.',
      { duration: 4000 }
    );
    return true;
  };

  /**
   * Ações de imagem (extrair / replicar / deletar) abrem o diálogo único da
   * aba (ImageActionDialogs) com a textura aberta como alvo — ali ficam a
   * escolha de alcance, o aviso de irreversível e a contagem de cópias.
   */
  const openAction = (type: 'extract' | 'replicate' | 'delete') => {
    // O hook é chamado antes do painel decidir que não há textura; sem alvo
    // não há ação a abrir.
    if (!entry) return;
    if (readOnly) {
      if (type === 'extract' && entry.vbf) {
        setBusy('extract');
        void extractVbfImagesSelection(entry.vbf.root, [entry.vbf.path], '')
          .then((result) => {
            if (result.failed.length > 0) {
              toast.warning(`Extraídas ${result.done.length} de ${result.total} imagens.`, {
                description: result.failed.slice(0, 3).join('; '),
              });
            } else if (result.total > 0) {
              toast.success('Imagem extraída como .dds e .png.');
            }
            if (result.done.length > 0 && entry.vbf) {
              invalidateVbfImage(entry.vbf.root, entry.vbf.path);
              void actions.selectEntry(entry);
            }
          })
          .catch((error) => toast.error(parseError(error)))
          .finally(() => setBusy(null));
        return;
      }
      if (type === 'replicate' && entry.vbf) {
        actions.openImageAction({
          type: 'replicate',
          id: entry.id,
          label: entry.label,
          duplicates,
          vbf: entry.vbf,
        });
        return;
      }
      refuseVbf();
      return;
    }
    actions.openImageAction({
      type,
      id: entry.id,
      label: entry.label,
      duplicates,
    });
  };

  const onSave = async (format: 'dds' | 'png') => {
    if (!entry) return;
    const name = entry.id.split('/').pop() ?? 'textura';
    setBusy('save');
    try {
      const dest = await selectImageSavePath(format, `${name}.${format}`, entry.id, version);
      if (!dest) return;
      if (entry.vbf) {
        await saveVbfImage(entry.vbf.root, entry.vbf.path, format, dest);
      } else {
        await saveImage(entry.id, format, dest, version);
      }
      toast.success(`Imagem salva: ${dest}`);
    } catch (error) {
      toast.error(parseError(error));
    } finally {
      setBusy(null);
    }
  };

  /**
   * Importa o .dds escolhido. `all=true` propaga para as cópias idênticas
   * (grupo de payload original — o backend recusa alvo fora dele).
   */
  const runImport = async (path: string, all: boolean) => {
    if (!entry) return;
    setBusy('import');
    try {
      if (entry.vbf) {
        await importVbfImage(entry.vbf.root, entry.vbf.path, path);
        invalidateVbfImage(entry.vbf.root, entry.vbf.path);
        toast.success(`Textura importada em mods/: ${entry.label}`);
        await actions.selectEntry(entry);
        return;
      }
      if (!all) {
        await importImage(entry.id, path, version);
        toast.success(`Textura importada: ${path}`);
      } else {
        const targets = duplicates.map((d) => d.id);
        const result = await importImageGroup(entry.id, path, targets, version);
        if (result.failed.length > 0) {
          toast.warning(
            `Importadas ${result.updated.length} de ${result.total} texturas.`,
            { description: result.failed.slice(0, 3).join('; ') }
          );
        } else {
          toast.success(
            `Textura importada em ${result.updated.length} textura(s) idêntica(s).`
          );
        }
      }
      // Recarrega árvore + textura: os containers em mods/ mudaram.
      await actions.reload();
    } catch (error) {
      toast.error(parseError(error));
    } finally {
      // O diálogo fica aberto durante a gravação (botões travados) e fecha
      // no fim — sucesso ou erro.
      setPendingImport(null);
      setBusy(null);
    }
  };

  const onImport = async () => {
    if (!entry) return;
    try {
      const path = await selectImageFile(entry.id, version);
      if (!path) return;
      if (readOnly) {
        await runImport(path, false);
        return;
      }
      // Com cópias idênticas o usuário escolhe o alcance antes de gravar.
      if (duplicates.length > 0) {
        setPendingImport(path);
        return;
      }
      await runImport(path, false);
    } catch (error) {
      toast.error(parseError(error));
    }
  };

  /** Reconstrói o índice de duplicatas (edição externa no hex editor). */
  const onRefresh = async () => {
    if (refuseVbf()) return;
    setBusy('refresh');
    try {
      await refreshImageDuplicates(version);
      toast.success('Duplicatas reanalisadas.');
      // Árvore junto: o índice é quem decide quais cópias ficam ocultas.
      await actions.reload();
    } catch (error) {
      toast.error(parseError(error));
    } finally {
      setBusy(null);
    }
  };

  return {
    busy,
    readOnly,
    pendingImport,
    setPendingImport,
    openAction,
    onSave,
    onImport,
    runImport,
    onRefresh,
  };
}
