'use client';

import { useEffect, useState } from 'react';
import { useSelector } from '@tanstack/react-store';
import { loggedToast as toast } from '@/lib/ffx/toast-logged';
import { AlertTriangle, Download, Info, Trash2 } from 'lucide-react';
import type { dto } from '@/wailsjs/go/models';
import { resolveEntryLabel } from '@/lib/ffx/display-names';
import { parseError } from '@/lib/ffx/error-handler';
import { exportSelection } from '@/lib/ffx/export-selection';
import {
  deleteImageSelection,
  deleteImages,
  extractImageGroup,
  imageDuplicates,
  imageSelectionCopies,
  replicateImage,
} from '@/lib/ffx/tree-data';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { copyState } from './image-duplicates';
import type { EntryView, ImageActionState } from './entry-view-store';
import { invalidateVbfImage, replicateVbfImage } from '@/lib/ffx/vbf';

type ActionProps = { view: EntryView; action: ImageActionState };
const EMPTY_IMAGE_IDS: string[] = [];

/**
 * Diálogos das ações de imagem (extrair / replicar / deletar).
 *
 * Um só ponto de renderização, na aba (game-version-tab): menu da árvore e
 * botões do painel apenas ABREM a ação no store — assim a ação serve ao nó
 * clicado, mesmo que não seja a entry selecionada, e não existe diálogo
 * duplicado por componente.
 */
export function ImageActionDialogs({ view }: { view: EntryView }) {
  const action = useSelector(view.store, (s) => s.imageAction);
  if (!action) return null;
  switch (action.type) {
    case 'extract':
      return <ExtractImageDialog view={view} action={action} />;
    case 'replicate':
      return <ReplicateImageDialog view={view} action={action} />;
    case 'delete-selection':
      return (
        <DeleteImageSelectionDialog
          key={`${view.version}:${action.ids?.join('\0') ?? ''}`}
          view={view}
          action={action}
        />
      );
    default:
      return <DeleteImageDialog view={view} action={action} />;
  }
}

/**
 * Cópias do alvo da ação: vêm prontas quando o chamador as conhecia (menu já
 * buscou / painel carrega a textura); null = carregando. Só busca quando não
 * vieram — o binding é leve, mas para quê chamá-lo?
 *
 * A lista guarda o id a que pertence: se a ação trocar de textura com o
 * diálogo montado, a lista antiga não serve (volta a "carregando").
 */
function useActionDuplicates(
  view: EntryView,
  action: ImageActionState
): dto.ImageDuplicate[] | null {
  const [state, setState] = useState<{ id: string; list: dto.ImageDuplicate[] | null }>(
    { id: action.id, list: action.duplicates }
  );

  useEffect(() => {
    // Veio pronta do chamador: nada a buscar.
    if (action.duplicates) return;
    // Lista desta textura já chegou (ou está a caminho).
    if (state.id === action.id && state.list !== null) return;
    let dead = false;
    imageDuplicates(action.id, view.version)
      .then((res) => {
        if (!dead) setState({ id: action.id, list: res.duplicates ?? [] });
      })
      .catch((error) => {
        if (dead) return;
        setState({ id: action.id, list: [] });
        toast.error(parseError(error));
      });
    return () => {
      dead = true;
    };
  }, [action.id, action.duplicates, state.id, state.list, view.version]);

  if (state.id !== action.id) return action.duplicates;
  return action.duplicates ?? state.list;
}

/** Lista da textura de referência + cópias, com o estado de cada uma. */
function DuplicateList({ id, list }: { id: string; list: dto.ImageDuplicate[] }) {
  return (
    <ul className="max-h-56 space-y-1 overflow-auto rounded border bg-muted/40 p-2 text-xs">
      <li className="flex items-center gap-1.5">
        <span className="min-w-0 flex-1 truncate">
          {resolveEntryLabel('images', id)}
        </span>
        <Badge variant="outline" className="text-[10px]">
          esta
        </Badge>
      </li>
      {list.map((d) => (
        <li key={d.id} className="flex items-center gap-1.5">
          <span className="min-w-0 flex-1 truncate">
            {resolveEntryLabel('images', d.id)}
          </span>
          <Badge variant={d.identical ? 'outline' : 'secondary'} className="text-[10px]">
            {copyState(d)}
          </Badge>
        </li>
      ))}
    </ul>
  );
}

function reportBatch(
  res: dto.BatchResult,
  singular: string,
  plural: string
): void {
  const label = res.done.length === 1 ? singular : plural;
  if (res.failed.length > 0) {
    toast.warning(`${label} ${res.done.length} de ${res.total} textura(s).`, {
      description: res.failed.slice(0, 3).join('; '),
    });
    return;
  }
  toast.success(
    `${label} ${res.done.length} textura${res.done.length === 1 ? '' : 's'}.`
  );
}

/** "Extrair só esta ou todas as cópias" — a confirmação de duplicatas. */
function ExtractImageDialog({ view, action }: ActionProps) {
  const { version, actions } = view;
  const list = useActionDuplicates(view, action);
  const [busy, setBusy] = useState(false);
  const copies = list ?? [];
  const divergent = copies.filter((d) => !d.identical).length;
  const identical = copies.length - divergent;
  const close = () => actions.closeImageAction();

  const run = async (targets: string[]) => {
    setBusy(true);
    try {
      const res = await extractImageGroup(action.id, targets, version);
      reportBatch(res, 'Extraída', 'Extraídas');
      await actions.reload();
      close();
    } catch (error) {
      toast.error(parseError(error));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !busy) close();
      }}
    >
      <DialogContent className="sm:max-w-lg" showCloseButton={!busy}>
        <DialogHeader>
          <DialogTitle>Extrair {action.label}?</DialogTitle>
          <DialogDescription>
            Gera .dds e .png em mods/edits/images (a cópia de trabalho, sem
            tocar em data/ nem em mods/*.dds.phyre). Escolha o alcance.
          </DialogDescription>
        </DialogHeader>

        {list === null ? (
          <p className="text-sm opacity-70">Carregando cópias…</p>
        ) : copies.length === 0 ? (
          <p className="text-sm opacity-70">
            Imagem única: não há cópias para extrair junto.
          </p>
        ) : (
          <>
            <p className="text-sm">
              Esta imagem tem {copies.length} cópia
              {copies.length > 1 ? 's' : ''}:{' '}
              {identical > 0
                ? `${identical} idêntica${identical > 1 ? 's' : ''}${divergent > 0 ? ' e ' : ''}`
                : ''}
              {divergent > 0
                ? `${divergent} com edição própria (divergente${divergent > 1 ? 's' : ''})`
                : ''}
              {' '}— extrair todas deixa os arquivos lado a lado para conferir
              a duplicata.
            </p>
            <DuplicateList id={action.id} list={copies} />
          </>
        )}

        <DialogFooter>
          <Button variant="outline" disabled={busy} onClick={close}>
            Cancelar
          </Button>
          <Button disabled={busy || list === null} onClick={() => void run([])}>
            <Download size={16} />
            {busy ? 'Extraindo…' : 'Extrair só esta'}
          </Button>
          {copies.length > 0 ? (
            <Button
              variant="outline"
              disabled={busy}
              onClick={() => void run(copies.map((d) => d.id))}
            >
              Extrair {copies.length + 1} texturas
            </Button>
          ) : null}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

/**
 * Replicar: leva a IMAGEM ABERTA para as cópias em mods/ — o dupe que o app
 * faz sozinho, sem diálogo de arquivo. As contagens avisam o que se perde
 * (divergentes sobrescritas) e o que se cria (pristine viram mods/).
 */
function ReplicateImageDialog({ view, action }: ActionProps) {
  const { version, actions } = view;
  const list = useActionDuplicates(view, action);
  const [busy, setBusy] = useState(false);
  const copies = list ?? [];
  const divergent = copies.filter((d) => !d.identical).length;
  const pristine = copies.filter((d) => !d.modded).length;
  const close = () => actions.closeImageAction();

  const run = async () => {
    setBusy(true);
    try {
      const res = action.vbf
        ? await replicateVbfImage(
            action.vbf.root,
            action.vbf.path,
            copies.map((d) => d.vbfPath ?? '').filter(Boolean)
          )
        : await replicateImage(
            action.id,
            copies.map((d) => d.id),
            version
          );
      reportBatch(res, 'Replicada', 'Replicadas');
      if (action.vbf) {
        invalidateVbfImage(action.vbf.root, action.vbf.path);
        const selected = view.store.state.selectedEntry;
        if (selected?.vbf?.root === action.vbf.root && selected.id === action.id) {
          await actions.selectEntry(selected);
        }
      } else {
        await actions.reload();
      }
      close();
    } catch (error) {
      toast.error(parseError(error));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !busy) close();
      }}
    >
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            Replicar {action.label}
            {list !== null && copies.length > 0
              ? ` para ${copies.length} cópia${copies.length === 1 ? '' : 's'}`
              : ''}
            ?
          </DialogTitle>
          <DialogDescription>
            {action.vbf
              ? 'A imagem lida do .vbf é reempacotada sobre o binário original de cada cópia e gravada em mods/. O container .vbf não é alterado.'
              : 'O conteúdo já carregado é reempacotado sobre o container pristine de cada cópia e gravado em mods/. Nenhum arquivo é escolhido: a fonte é a própria imagem na tela, e a textura de origem não é tocada.'}
          </DialogDescription>
        </DialogHeader>

        {list === null ? (
          <p className="text-sm opacity-70">Carregando cópias…</p>
        ) : copies.length === 0 ? (
          <p className="text-sm opacity-70">
            Imagem única: não há cópias para replicar.
          </p>
        ) : (
          <>
            {divergent > 0 ? (
              <Alert variant="warning">
                <AlertTriangle />
                <AlertTitle>
                  {divergent} cópia{divergent > 1 ? 's' : ''} divergente
                  {divergent > 1 ? 's' : ''} serão sobrescrita
                  {divergent > 1 ? 's' : ''}
                </AlertTitle>
                <AlertDescription>
                  Foram editadas em separado e perderão a edição atual.
                </AlertDescription>
              </Alert>
            ) : null}
            {pristine > 0 ? (
              <Alert variant="info">
                <Info />
                <AlertTitle>
                  {pristine} cópia{pristine > 1 ? 's' : ''} pristine receber
                  {pristine > 1 ? 'ão' : ''} arquivo em mods/
                </AlertTitle>
                <AlertDescription>
                  data/ continua intacto — dá para desfazer apagando só mods/.
                </AlertDescription>
              </Alert>
            ) : null}
            <DuplicateList id={action.id} list={copies} />
          </>
        )}

        <DialogFooter>
          <Button variant="outline" disabled={busy} onClick={close}>
            Cancelar
          </Button>
          <Button
            disabled={busy || list === null || copies.length === 0}
            onClick={() => void run()}
          >
            {busy
              ? 'Replicando…'
              : list === null
                ? 'Carregando…'
                : `Replicar para ${copies.length} cópia${
                    copies.length === 1 ? '' : 's'
                  }`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

/** Escopos do delete: onde os containers são apagados. */
const DELETE_SCOPES: Array<{
  value: 'both' | 'data' | 'mods';
  title: string;
  hint: string;
}> = [
  {
    value: 'both',
    title: 'Ambos (data/ e mods/)',
    hint: 'apaga o original e a substituição — a textura sai da árvore',
  },
  {
    value: 'data',
    title: 'Só o original (data/)',
    hint: 'a textura sai da árvore; mods/ permanece como está',
  },
  {
    value: 'mods',
    title: 'Só em mods/',
    hint: 'desfaz a substituição e volta ao conteúdo original',
  },
];

/**
 * Deletar: escopo explícito (nunca cascata implícita) + cópias só se
 * marcadas, com o aviso de irreversível antes de qualquer gravação. Sempre
 * leva junto os .dds/.png extraídos — senão Resolve serviria imagem órfã.
 */
function DeleteImageDialog({ view, action }: ActionProps) {
  const { version, actions } = view;
  const list = useActionDuplicates(view, action);
  const [scope, setScope] = useState<'both' | 'data' | 'mods'>('both');
  const [withCopies, setWithCopies] = useState(false);
  const [busy, setBusy] = useState(false);
  const copies = list ?? [];
  const targets = withCopies ? copies.map((d) => d.id) : [];
  const divergent = copies.filter((d) => !d.identical).length;
  const close = () => actions.closeImageAction();

  const run = async () => {
    setBusy(true);
    try {
      const res = await deleteImages(action.id, targets, scope, version);
      reportBatch(res, 'Apagada', 'Apagadas');
      await actions.reload();
      close();
    } catch (error) {
      toast.error(parseError(error));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !busy) close();
      }}
    >
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Deletar {action.label}?</DialogTitle>
          <DialogDescription>
            Escolha o escopo. Os .dds/.png extraídos desta textura são
            apagados junto em qualquer escopo.
          </DialogDescription>
        </DialogHeader>

        <Alert variant="destructive">
          <AlertTriangle />
          <AlertTitle>Irreversível</AlertTitle>
          <AlertDescription>
            Os arquivos são removidos do disco, fora da lixeira, e não há
            desfazer — nem o app consegue reconstruí-los sem extrair o
            FFX_Data.vbf de novo.
          </AlertDescription>
        </Alert>

        <fieldset className="space-y-1.5">
          <legend className="mb-1 text-xs uppercase tracking-wide opacity-60">
            Escopo
          </legend>
          {DELETE_SCOPES.map((option) => (
            <label
              key={option.value}
              className={`flex cursor-pointer items-start gap-2 rounded border p-2 text-sm ${
                scope === option.value
                  ? 'border-primary bg-primary/5'
                  : 'border-border'
              }`}
            >
              <input
                type="radio"
                name="image-delete-scope"
                className="mt-0.5"
                checked={scope === option.value}
                onChange={() => setScope(option.value)}
              />
              <span className="min-w-0">
                <span className="block font-medium">{option.title}</span>
                <span className="block text-xs opacity-70">{option.hint}</span>
              </span>
            </label>
          ))}
        </fieldset>

        {copies.length > 0 ? (
          <>
            <div className="flex items-center gap-2">
              <Checkbox
                id="delete-copies"
                checked={withCopies}
                onCheckedChange={(value) => setWithCopies(value === true)}
              />
              <Label htmlFor="delete-copies" className="cursor-pointer text-sm">
                Apagar também as {copies.length} cópia
                {copies.length > 1 ? 's' : ''}
              </Label>
            </div>
            {withCopies && divergent > 0 ? (
              <Alert variant="warning">
                <AlertTriangle />
                <AlertTitle>
                  {divergent} cópia{divergent > 1 ? 's' : ''} diverge
                  {divergent > 1 ? 'm' : ''} do original
                </AlertTitle>
                <AlertDescription>
                  Estão divergentes porque foram editadas em separado — a
                  edição delas também será apagada.
                </AlertDescription>
              </Alert>
            ) : null}
          </>
        ) : (
          <p className="text-sm opacity-70">Imagem única: sem cópias.</p>
        )}

        <DialogFooter>
          <Button variant="outline" disabled={busy} onClick={close}>
            Cancelar
          </Button>
          <Button
            variant="destructive"
            disabled={busy || list === null}
            onClick={() => void run()}
          >
            <Trash2 size={16} />
            {busy
              ? 'Apagando…'
              : `Deletar ${targets.length + 1} textura${
                  targets.length > 0 ? 's' : ''
                }`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function DeleteImageSelectionDialog({ view, action }: ActionProps) {
  const { version, actions } = view;
  const ids = action.ids ?? EMPTY_IMAGE_IDS;
  const [scope, setScope] = useState<'both' | 'data' | 'mods'>('both');
  const [withCopies, setWithCopies] = useState(false);
  const [busy, setBusy] = useState(false);
  const [copyIDs, setCopyIDs] = useState<string[]>([]);
  const [duplicatesReady, setDuplicatesReady] = useState(false);
  const [copyLookupError, setCopyLookupError] = useState<string | null>(null);
  const close = () => actions.closeImageAction();

  useEffect(() => {
    let cancelled = false;
    void imageSelectionCopies(ids, version)
      .then((copies) => {
        if (cancelled) return;
        setCopyIDs(copies ?? []);
        setDuplicatesReady(true);
      })
      .catch((error) => {
        if (cancelled) return;
        setCopyIDs([]);
        setCopyLookupError(parseError(error));
        setDuplicatesReady(true);
        toast.error(parseError(error));
      });
    return () => {
      cancelled = true;
    };
  }, [ids, version]);

  const targets = withCopies ? copyIDs : [];
  const deleteCount = ids.length + targets.length;
  const run = async () => {
    if (ids.length === 0) return;
    setBusy(true);
    try {
      const res = await deleteImageSelection(ids, withCopies, scope, version);
      reportBatch(res, 'Apagada', 'Apagadas');
      exportSelection.setMany(version, 'images', res.done ?? [], false);
      await actions.reload();
      close();
    } catch (error) {
      toast.error(parseError(error));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !busy) close();
      }}
    >
      <DialogContent className="sm:max-w-lg" showCloseButton={!busy}>
        <DialogHeader>
          <DialogTitle>
            Deletar {ids.length} {ids.length === 1 ? 'imagem selecionada' : 'imagens selecionadas'}?
          </DialogTitle>
          <DialogDescription>
            O escopo escolhido será aplicado a todas as imagens selecionadas.
            As cópias só serão incluídas se a opção abaixo estiver marcada.
          </DialogDescription>
        </DialogHeader>

        <Alert variant="destructive">
          <AlertTriangle />
          <AlertTitle>Irreversível</AlertTitle>
          <AlertDescription>
            Os arquivos são removidos do disco, fora da lixeira, e não há
            desfazer. Os `.dds` e `.png` extraídos também serão apagados.
          </AlertDescription>
        </Alert>

        <fieldset className="space-y-1.5">
          <legend className="mb-1 text-xs uppercase tracking-wide opacity-60">
            Escopo para todas as selecionadas
          </legend>
          {DELETE_SCOPES.map((option) => (
            <label
              key={option.value}
              className={`flex cursor-pointer items-start gap-2 rounded border p-2 text-sm ${
                scope === option.value
                  ? 'border-primary bg-primary/5'
                  : 'border-border'
              }`}
            >
              <input
                type="radio"
                name="image-delete-selection-scope"
                className="mt-0.5"
                checked={scope === option.value}
                onChange={() => setScope(option.value)}
              />
              <span className="min-w-0">
                <span className="block font-medium">{option.title}</span>
                <span className="block text-xs opacity-70">{option.hint}</span>
              </span>
            </label>
          ))}
        </fieldset>

        {duplicatesReady ? (
          copyLookupError ? (
            <Alert variant="warning">
              <AlertTriangle />
              <AlertTitle>Não foi possível verificar as cópias</AlertTitle>
              <AlertDescription>
                A confirmação ainda pode apagar somente as imagens marcadas.
                Cópias adicionais não serão incluídas.
              </AlertDescription>
            </Alert>
          ) : copyIDs.length > 0 ? (
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Checkbox
                  id="delete-selection-copies"
                  checked={withCopies}
                  onCheckedChange={(value) => setWithCopies(value === true)}
                />
                <Label htmlFor="delete-selection-copies" className="cursor-pointer text-sm">
                  Apagar também {copyIDs.length} cópia
                  {copyIDs.length === 1 ? '' : 's'} adicional
                  {copyIDs.length === 1 ? '' : 'is'}
                </Label>
              </div>
              {withCopies ? (
                <Alert variant="warning">
                  <AlertTriangle />
                  <AlertTitle>
                    {targets.length} cópia(s) adicionais serão incluídas
                  </AlertTitle>
                  <AlertDescription>
                    Cópias podem ter sido editadas separadamente. Todas as
                    cópias adicionais serão apagadas com o mesmo escopo.
                  </AlertDescription>
                </Alert>
              ) : null}
            </div>
          ) : (
            <p className="text-sm opacity-70">As imagens selecionadas não têm cópias.</p>
          )
        ) : (
          <p className="text-sm opacity-70">Verificando cópias das selecionadas…</p>
        )}

        <DialogFooter>
          <Button variant="outline" disabled={busy} onClick={close}>
            Cancelar
          </Button>
          <Button
            variant="destructive"
            disabled={busy || !duplicatesReady || ids.length === 0}
            onClick={() => void run()}
          >
            <Trash2 size={16} />
            {busy
              ? 'Apagando…'
              : `Deletar ${deleteCount} ${deleteCount === 1 ? 'imagem' : 'imagens'}`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
