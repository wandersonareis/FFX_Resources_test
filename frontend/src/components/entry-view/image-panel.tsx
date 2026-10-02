'use client';

import { useState, type ReactNode } from 'react';
import { useSelector } from '@tanstack/react-store';
import { toast } from 'sonner';
import {
  Copy,
  Download,
  FolderInput,
  FlipVertical,
  RefreshCw,
  Save,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { parseError } from '@/lib/ffx/error-handler';
import { resolveEntryLabel } from '@/lib/ffx/display-names';
import {
  importImage,
  importImageGroup,
  refreshImageDuplicates,
  saveImage,
  selectImageFile,
  selectImageSavePath,
  type EntryRow,
} from '@/lib/ffx/tree-data';
import { EntryActionsMenu, type EntryMenuTarget } from './entry-actions-menu';
import { copyState } from './image-duplicates';
import type { EntryView } from './entry-view-store';

/** De onde veio a imagem servida (o backend decide, por preferência). */
const SOURCE_LABELS: Record<string, string> = {
  dds: '.dds em disco',
  png: '.png em disco',
  phyre: 'decodificado do .dds.phyre',
};

/** Tamanho legível para o desperdício do grupo de cópias. */
function bytesLabel(bytes: number): string {
  if (bytes >= 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  if (bytes >= 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${bytes} B`;
}

/**
 * Painel da textura (kind=images): ocupa o lugar da tabela Original/
 * Traduzido e mostra a pré-visualização PNG que o backend já devolve
 * (data URL — nenhum decoder de imagem no JS).
 *
 * Ações: extrair cópias de trabalho (.dds + .png em mods/edits/images),
 * salvar em qualquer lugar do disco, importar um .dds editado (repack só do
 * mesmo formato/dimensão nesta etapa) e, quando a imagem tem cópias
 * idênticas (otimização do DVD), propagar o import para elas.
 *
 * As duplicatas vêm do índice de duplicatas do backend (hash do PAYLOAD, não
 * do arquivo: o namespace embute o nome e tornaria inútil o hash do
 * arquivo). O índice é memoizado; "Reanalisar" o reconstrói para pegar
 * edição externa em hex editor.
 */
export function ImagePanel({ view }: { view: EntryView }) {
  const { version, store, actions } = view;
  const entry = useSelector(store, (s) => s.selectedEntry);
  const image = useSelector(store, (s) => s.image);
  const loading = useSelector(store, (s) => s.loading);
  const [busy, setBusy] = useState<'import' | 'save' | 'refresh' | null>(null);
  /**
   * Textura com o preview espelhado (id) — SÓ VISUAL, um transform no
   * <img>: nenhum byte muda e nada é mandado de volta ao backend, que já
   * entrega a imagem na orientação correta (flip feito ao decodificar).
   * Guarda o id (não um booleano) para cada textura começar de pé — sem
   * useEffect de sincronização.
   */
  const [flippedId, setFlippedId] = useState<string | null>(null);
  /**
   * Zoom manual do preview: id da textura + nível ('fit' = encaixar no
   * painel). Guarda o id para cada textura voltar ao ajuste automático,
   * no mesmo padrão do flippedId (sem useEffect). O nível é o multiplicador
   * do tamanho NATIVO da textura (1× = 1 pixel da textura por pixel da tela).
   */
  const [zoom, setZoom] = useState<{ id: string; level: 'fit' | number } | null>(
    null,
  );
  /**
   * Renderização pixelada (nearest-neighbor) — preferência da sessão. É o
   * jeito de distinguir embaçamento da EXIBIÇÃO (escala bilinear do browser
   * ao encaixar texturas pequenas no painel) do embaçamento do DADO (DXT
   * comprimido / arte de baixa resolução): em 1× com pixelado, o que se vê
   * é exatamente o que o decoder do jogo produz.
   */
  const [pixelated, setPixelated] = useState(false);
  /**
   * .dds já escolhido aguardando o usuário decidir o alcance do import
   * (null = sem diálogo aberto). Só aparece quando a imagem tem cópias.
   */
  const [pendingImport, setPendingImport] = useState<string | null>(null);
  /**
   * Menu de contexto do painel (o mesmo da árvore). O alvo é SEMPRE a
   * textura aberta — todo o painel é o nó — então o target é fixo quando
   * aberto e null quando fechado.
   */
  const [menuTarget, setMenuTarget] = useState<EntryMenuTarget | null>(null);

  if (!entry) return null;
  const flipped = flippedId === entry.id;
  const zoomLevel = zoom?.id === entry.id ? zoom.level : 'fit';
  if (loading && !image) return <p className="mt-8 opacity-70">Carregando…</p>;
  if (!image) {
    return (
      <p className="mt-8 opacity-70">Não foi possível carregar a textura.</p>
    );
  }

  const name = entry.id.split('/').pop() ?? 'textura';
  const duplicates = image.duplicates ?? [];
  const wasted = image.dupPayload * duplicates.length;

  /** Navega para uma cópia (a árvore já tem o entry — é só selecionar). */
  const goToCopy = async (id: string, key: string): Promise<void> => {
    const target: EntryRow = {
      kind: 'images',
      id,
      key,
      label: resolveEntryLabel('images', id),
    };
    await actions.selectEntry(target);
  };

  /**
   * Ações de imagem (extrair / replicar / deletar) abrem o diálogo único da
   * aba (ImageActionDialogs) com a textura aberta como alvo — ali ficam a
   * escolha de alcance, o aviso de irreversível e a contagem de cópias.
   */
  const openAction = (type: 'extract' | 'replicate' | 'delete') => {
    actions.openImageAction({
      type,
      id: entry.id,
      label: entry.label,
      duplicates,
    });
  };

  const onSave = async (format: 'dds' | 'png') => {
    setBusy('save');
    try {
      const dest = await selectImageSavePath(format, `${name}.${format}`);
      if (!dest) return;
      await saveImage(entry.id, format, dest, version);
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
    setBusy('import');
    try {
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
    try {
      const path = await selectImageFile();
      if (!path) return;
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

  // Linha "Repetidas": o grupo de cópias de mesma imagem (otimização do
  // DVD) com atalho de navegação para cada uma delas.
  const repeatsRow: [label: string, value: ReactNode] =
    duplicates.length > 0
      ? [
          'Repetidas',
          <div key="dups" className="space-y-1.5">
            <div className="flex flex-wrap items-center gap-1.5">
              <Badge className="bg-amber-600 text-white">
                <Copy size={12} className="mr-1" />×{duplicates.length + 1}{' '}
                idênticas
              </Badge>
              <span className="text-xs opacity-70">
                {bytesLabel(wasted)} de cópias
              </span>
            </div>
            <ul className="space-y-0.5">
              {duplicates.map((d) => (
                <li key={d.id}>
                  <button
                    type="button"
                    className="flex w-full items-center gap-1.5 rounded px-1 py-0.5 text-left text-xs hover:bg-muted"
                    title={`${d.id} — abrir esta cópia`}
                    onClick={() => void goToCopy(d.id, d.key)}
                  >
                    <span className="min-w-0 flex-1 truncate underline-offset-2 hover:underline">
                      {resolveEntryLabel('images', d.id)}
                    </span>
                    <Badge
                      variant={d.identical ? 'outline' : 'secondary'}
                      className="shrink-0 text-[10px]"
                    >
                      {copyState(d)}
                    </Badge>
                  </button>
                </li>
              ))}
            </ul>
          </div>,
        ]
      : [
          'Repetidas',
          <span key="dups" className="opacity-70">
            nenhuma (imagem única)
          </span>,
        ];

  const meta: Array<[label: string, value: ReactNode]> = [
    ['Formato', image.format],
    ['Dimensões', `${image.width}×${image.height}`],
    [
      'Mipmaps',
      `${image.mipmapCount}${image.maxMipmapLevel ? ` (max ${image.maxMipmapLevel})` : ''}`,
    ],
    [
      'Fonte',
      <Badge key="source" variant="outline">
        {SOURCE_LABELS[image.source] ?? image.source}
      </Badge>,
    ],
    [
      'Estado',
      image.modded ? (
        <Badge key="modded" className="bg-emerald-600 text-white">
          importada (mods/)
        </Badge>
      ) : (
        <Badge key="modded" variant="outline">
          pristine (data/)
        </Badge>
      ),
    ],
    repeatsRow,
    ['Cópias extraídas', image.ddsPath || image.pngPath ? (
      <span key="paths" className="block break-all opacity-80">
        {image.ddsPath ? <span className="block">{image.ddsPath}</span> : null}
        {image.pngPath ? <span className="block">{image.pngPath}</span> : null}
      </span>
    ) : (
      'nenhuma'
    )],
  ];

  const copyTotal = duplicates.length + 1;

  // Alvo do menu de contexto: o painel inteiro é a textura aberta.
  const panelTarget: EntryMenuTarget = {
    kind: 'images',
    id: entry.id,
    label: entry.label,
  };

  // Botão direito em QUALQUER ponto do painel abre o mesmo menu da árvore
  // (Exportar / Abrir até o arquivo / Replicar / Deletar).
  return (
    <EntryActionsMenu
      view={view}
      target={menuTarget}
      onOpenChange={(open) => setMenuTarget(open ? panelTarget : null)}
      initialId={entry.id}
      initialDuplicates={duplicates}
    >
      <div
        className="mt-4 flex flex-col gap-4"
        onContextMenuCapture={() => setMenuTarget(panelTarget)}
      >
        <div className="flex flex-wrap items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            disabled={busy !== null}
            title="Gera .dds e .png em mods/edits/images — o diálogo pergunta se é só esta ou todas as cópias"
            onClick={() => openAction('extract')}
          >
            <Download size={16} />
            Extrair .dds + .png
          </Button>
          <Button
            size="sm"
            variant="outline"
            disabled={busy !== null}
            onClick={() => void onSave('dds')}
          >
            <Save size={16} />
            Salvar .dds
          </Button>
          <Button
            size="sm"
            variant="outline"
            disabled={busy !== null}
            onClick={() => void onSave('png')}
          >
            <Save size={16} />
            Salvar .png
          </Button>
          <Button
            size="sm"
            disabled={busy !== null}
            onClick={() => void onImport()}
          >
            <FolderInput size={16} />
            {busy === 'import'
              ? 'Importando…'
              : duplicates.length > 0
                ? `Importar .dds… (${copyTotal} idênticas)`
                : 'Importar .dds…'}
          </Button>
          {duplicates.length > 0 ? (
            <Button
              size="sm"
              variant="outline"
              disabled={busy !== null}
              title="Reempacota a imagem aberta em mods/ de cada cópia idêntica — sem escolher arquivo"
              onClick={() => openAction('replicate')}
            >
              <Copy size={16} />
              Replicar para {duplicates.length} cópia
              {duplicates.length > 1 ? 's' : ''}
            </Button>
          ) : null}
          <Button
            size="sm"
            variant={flipped ? 'default' : 'outline'}
            aria-pressed={flipped}
            title="Espelha a pré-visualização — só na tela, não altera nem reenvia a imagem"
            onClick={() => setFlippedId(flipped ? null : entry.id)}
          >
            <FlipVertical size={16} />
            {flipped ? 'Flip (on)' : 'Flip'}
          </Button>
          <Button
            size="sm"
            variant={zoomLevel === 'fit' ? 'default' : 'outline'}
            aria-pressed={zoomLevel === 'fit'}
            title="Encaixa a textura no painel (escala suave do navegador)"
            onClick={() => setZoom(null)}
          >
            Ajustar
          </Button>
          {[1, 2, 4].map((level) => (
            <Button
              key={level}
              size="sm"
              variant={zoomLevel === level ? 'default' : 'outline'}
              aria-pressed={zoomLevel === level}
              title={`Zoom ${level}× — 1 pixel da textura = ${level} na tela. Em 1× o que se vê é exatamente o dado; nítido aqui = o embaçado no Ajustar era só a escala`}
              onClick={() => setZoom({ id: entry.id, level })}
            >
              {level}×
            </Button>
          ))}
          <Button
            size="sm"
            variant={pixelated ? 'default' : 'outline'}
            aria-pressed={pixelated}
            title="Renderização pixelada (nearest-neighbor) — cada pixel como é, sem suavização do navegador"
            onClick={() => setPixelated(!pixelated)}
          >
            Pixelado
          </Button>
          <Button
            size="sm"
            variant="ghost"
            disabled={busy !== null}
            title="Reconstrói o índice de duplicatas — para textura editada fora do app (hex editor)"
            onClick={() => void onRefresh()}
          >
            <RefreshCw size={16} />
            {busy === 'refresh' ? 'Reanalisando…' : 'Reanalisar'}
          </Button>
        </div>

        <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_20rem]">
          <div className="flex min-h-40 items-center justify-center overflow-auto rounded-md border bg-muted/40 p-4">
            {image.pngData ? (
              // eslint-disable-next-line @next/next/no-img-element -- data URL gerado pelo Go (next/image não otimiza nem precisa)
              <img
                src={image.pngData}
                alt={entry.label}
                // Só visual: o backend já envia na orientação do jogo; o
                // espelhamento, o zoom e o pixelado não tocam no dado nem
                // geram chamada ao Go. No zoom o scroll do container navega.
                style={{
                  transform: flipped ? 'scaleY(-1)' : undefined,
                  width:
                    zoomLevel === 'fit' ? undefined : `${image.width * zoomLevel}px`,
                  imageRendering: pixelated ? 'pixelated' : undefined,
                  transition: 'transform 120ms ease-out',
                }}
                className={
                  zoomLevel === 'fit'
                    ? 'max-h-[65vh] max-w-full object-contain'
                    : undefined
                }
              />
            ) : (
              <p className="opacity-70">Sem pré-visualização disponível.</p>
            )}
          </div>

          <dl className="space-y-2 text-sm">
            {meta.map(([label, value]) => (
              <div key={label} className="flex flex-col gap-1">
                <dt className="text-xs uppercase tracking-wide opacity-60">
                  {label}
                </dt>
                <dd className="min-w-0">{value}</dd>
              </div>
            ))}
          </dl>
        </div>

        <Dialog
          open={pendingImport !== null}
          onOpenChange={(open) => {
            if (!open) setPendingImport(null);
          }}
        >
          <DialogContent className="sm:max-w-md">
            <DialogHeader>
              <DialogTitle>Importar em {copyTotal} texturas?</DialogTitle>
              <DialogDescription>
                Esta imagem tem {duplicates.length} cópias idênticas — a
                otimização do DVD repetiu o mesmo conteúdo por caminho. O mesmo
                .dds pode ser reempacotado sobre todas para mantê-las
                sincronizadas. Cópias divergentes serão sobrescritas.
              </DialogDescription>
            </DialogHeader>

            <ul className="max-h-56 space-y-1 overflow-auto rounded border bg-muted/40 p-2 text-xs">
              <li className="flex items-center gap-1.5">
                <span className="min-w-0 flex-1 truncate">
                  {resolveEntryLabel('images', entry.id)}
                </span>
                <Badge variant="outline" className="text-[10px]">
                  esta
                </Badge>
              </li>
              {duplicates.map((d) => (
                <li key={d.id} className="flex items-center gap-1.5">
                  <span className="min-w-0 flex-1 truncate">
                    {resolveEntryLabel('images', d.id)}
                  </span>
                  <Badge
                    variant={d.identical ? 'outline' : 'secondary'}
                    className="text-[10px]"
                  >
                    {copyState(d)}
                  </Badge>
                </li>
              ))}
            </ul>

            <DialogFooter>
              <Button
                variant="outline"
                disabled={busy !== null}
                onClick={() => setPendingImport(null)}
              >
                Cancelar
              </Button>
              <Button
                variant="outline"
                disabled={busy !== null}
                onClick={() => pendingImport && void runImport(pendingImport, false)}
              >
                Só nesta
              </Button>
              <Button
                disabled={busy !== null}
                onClick={() => pendingImport && void runImport(pendingImport, true)}
              >
                <Copy size={16} className="mr-1" />
                Importar em {copyTotal}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </EntryActionsMenu>
  );
}
