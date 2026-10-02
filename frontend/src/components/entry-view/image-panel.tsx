'use client';

import { useState, type ReactNode } from 'react';
import { useHotkey } from '@tanstack/react-hotkeys';
import { useSelector } from '@tanstack/react-store';
import { toast } from 'sonner';
import {
  ChevronDown,
  Copy,
  Download,
  FolderInput,
  FlipVertical,
  Maximize2,
  RefreshCw,
  Save,
  Trash2,
  ZoomIn,
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
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
 * Níveis discretos do ciclo da lupa: tamanho real (1×) até 4× — depois de
 * 4× volta ao tamanho real. O slider permite valores intermediários.
 */
const ZOOM_LEVELS = [1, 2, 3, 4];

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

  const zoomLevel = zoom && zoom.id === entry?.id ? zoom.level : 'fit';
  /** Lupa: próximo nível do ciclo — depois de 4× volta ao tamanho real. */
  const zoomIn = () => {
    const current = zoomLevel === 'fit' ? 0 : zoomLevel;
    setZoom({
      id: entry?.id ?? '',
      level: ZOOM_LEVELS.find((l) => l > current) ?? ZOOM_LEVELS[0],
    });
  };
  /** Passo atrás no ciclo; sem efeito no ajuste automático. */
  const zoomOut = () => {
    if (zoomLevel === 'fit') return;
    const prev = [...ZOOM_LEVELS].reverse().find((l) => l < zoomLevel);
    if (prev) setZoom({ id: entry?.id ?? '', level: prev });
  };

  // Zoom por teclado (TanStack Hotkeys): Mod+= / Mod+- andam no ciclo e
  // Mod+0 volta ao ajuste; 1–4 vão direto ao nível SEM Mod (Mod+1/2/3 já
  // trocam de aba no app-shell). Dígito só dispara fora de inputs e com o
  // painel de textura na frente.
  const zoomKeysEnabled = entry !== null && pendingImport === null;
  useHotkey('Mod+=', zoomIn, { enabled: zoomKeysEnabled, ignoreInputs: true });
  useHotkey('Mod+-', zoomOut, { enabled: zoomKeysEnabled, ignoreInputs: true });
  useHotkey('Mod+0', () => setZoom(null), {
    enabled: zoomKeysEnabled,
    ignoreInputs: true,
  });
  useHotkey('1', () => setZoom({ id: entry?.id ?? '', level: 1 }), {
    enabled: zoomKeysEnabled,
    ignoreInputs: true,
  });
  useHotkey('2', () => setZoom({ id: entry?.id ?? '', level: 2 }), {
    enabled: zoomKeysEnabled,
    ignoreInputs: true,
  });
  useHotkey('3', () => setZoom({ id: entry?.id ?? '', level: 3 }), {
    enabled: zoomKeysEnabled,
    ignoreInputs: true,
  });
  useHotkey('4', () => setZoom({ id: entry?.id ?? '', level: 4 }), {
    enabled: zoomKeysEnabled,
    ignoreInputs: true,
  });

  if (!entry) return null;
  const flipped = flippedId === entry.id;
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
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                size="sm"
                variant="outline"
                disabled={busy !== null}
                title="Salvar como — o .dds é o padrão (sem perda, volta no repack); o .png é só para visualizar/distribuir"
              >
                <Save size={16} />
                Exportar
                <ChevronDown size={12} />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
              <DropdownMenuItem onClick={() => void onSave('dds')}>
                <Save size={14} />
                .dds
                <span className="ml-2 text-xs opacity-60">padrão</span>
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => void onSave('png')}>
                <Download size={14} />
                .png
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
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
            variant="outline"
            disabled={busy !== null}
            title="Apaga a textura conforme o escopo (data/mods/ambos) — o diálogo pede confirmação"
            onClick={() => openAction('delete')}
          >
            <Trash2 size={16} />
            Deletar
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
          <div className="relative">
            {/* Área de exibição com altura fixa: no zoom a imagem cresce
                além do container e rola em ambas as direções (comportamento
                de visualizador de imagem na web), por escolha do usuário. */}
            <div
              className={
                zoomLevel === 'fit'
                  ? 'flex min-h-40 items-center justify-center overflow-auto rounded-md border bg-muted/40 p-4'
                  : 'flex h-[65vh] overflow-auto rounded-md border bg-muted/40 p-4'
              }
            >
            {image.pngData ? (
              // eslint-disable-next-line @next/next/no-img-element -- data URL gerado pelo Go (next/image não otimiza nem precisa)
              <img
                src={image.pngData}
                alt={entry.label}
                // Só visual: o backend já envia na orientação do jogo; o
                // espelhamento, o zoom e o pixelado não tocam no dado nem
                // geram chamada ao Go. No zoom o scroll do container navega.
                // Escala UNIFORME explícita: largura e altura multiplicadas
                // pelo mesmo nível — o aspecto nunca muda, nem em 1× (tamanho
                // real). maxWidth: 'none' anula o preflight do Tailwind
                // (img { max-width: 100%; height: auto }), que reencaparía a
                // imagem de volta ao container e quebraria a rolagem.
                style={{
                  transform: flipped ? 'scaleY(-1)' : undefined,
                  width:
                    zoomLevel === 'fit' ? undefined : `${image.width * zoomLevel}px`,
                  height:
                    zoomLevel === 'fit' ? undefined : `${image.height * zoomLevel}px`,
                  maxWidth: zoomLevel === 'fit' ? undefined : 'none',
                  imageRendering: pixelated ? 'pixelated' : undefined,
                  transition: 'transform 120ms ease-out',
                }}
                className={
                  zoomLevel === 'fit'
                    ? 'max-h-[65vh] max-w-full object-contain'
                    : // m-auto centraliza quando cabe e zera quando estoura —
                      // sem isso o flex centrado cortaria a rolagem; shrink-0
                      // impede o flex de encolher a imagem de volta.
                      'm-auto shrink-0'
                }
              />
            ) : (
              <p className="opacity-70">Sem pré-visualização disponível.</p>
            )}
            </div>
            {/* Controles de zoom flutuando sobre a imagem: a lupa anda no
                ciclo 1× (tamanho real) → 4× e volta; ajuste fino no slider. */}
            <div className="pointer-events-none absolute inset-x-0 top-0 z-10 flex flex-col items-end gap-1 p-2">
              <div className="pointer-events-auto flex items-center gap-1 rounded-md border bg-background/85 p-1 shadow-sm backdrop-blur">
                <Button
                  size="icon"
                  variant="ghost"
                  className="size-7"
                  title="Lupa: tamanho real → 2× → 3× → 4× → tamanho real (teclado: Mod+= / Mod+-; Mod+0 volta ao ajuste)"
                  onClick={zoomIn}
                >
                  <ZoomIn size={16} />
                </Button>
                <Button
                  size="sm"
                  variant={zoomLevel === 'fit' ? 'default' : 'ghost'}
                  className="h-7 px-2 text-xs"
                  title="Encaixa a textura no painel (escala suave do navegador)"
                  onClick={() => setZoom(null)}
                >
                  <Maximize2 size={14} />
                  Ajustar
                </Button>
                <Button
                  size="sm"
                  variant={pixelated ? 'default' : 'ghost'}
                  className="h-7 px-2 text-xs"
                  title="Pixelado (nearest-neighbor) — cada pixel como é; distingue embaçamento da exibição do embaçamento do dado"
                  onClick={() => setPixelated(!pixelated)}
                >
                  Pixelado
                </Button>
              </div>
              {zoomLevel !== 'fit' ? (
                <div className="pointer-events-auto flex items-center gap-2 rounded-md border bg-background/85 px-2 py-1 shadow-sm backdrop-blur">
                  <span className="text-xs opacity-60">real</span>
                  <input
                    type="range"
                    min={1}
                    max={4}
                    step={0.25}
                    value={zoomLevel}
                    aria-label="Zoom da pré-visualização"
                    className="w-36 accent-primary"
                    onChange={(e) =>
                      setZoom({ id: entry.id, level: Number(e.target.value) })
                    }
                  />
                  <span className="w-9 text-right text-xs tabular-nums opacity-60">
                    {zoomLevel}×
                  </span>
                </div>
              ) : null}
            </div>
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
