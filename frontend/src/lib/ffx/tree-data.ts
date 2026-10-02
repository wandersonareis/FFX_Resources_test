import {
  DeleteImages,
  ExportEntry,
  ExportJSON,
  ExportStrings,
  ExtractImageGroup,
  GetImageEntry,
  GetTextEntry,
  ImageDuplicates,
  ImageExists,
  ImportEntry,
  ImportFile,
  ImportImage,
  ImportImageGroup,
  ListTextEntries,
  PreviewImport,
  RefreshImageDuplicates,
  ReplicateImage,
  RevealEntryFile,
  SaveImage,
  SelectImageFile,
  SelectImageSavePath,
  SelectImportFile,
} from '@/wailsjs/go/main/App';
import { dto, services } from '@/wailsjs/go/models';
import { resolveEntryLabel, EntryKind } from './display-names';
import type { GameVersionId } from './game-version';

export interface EntryRow {
  kind: EntryKind;
  id: string;
  key: string;
  label: string;
}

// Kinds servidos por versão na sidebar. lastmiss é expansão do ffx2 e não tem
// dicionário próprio → sem 'macro' (a aba não mostra "Dicionário"). O lockit
// (kit de localização do menu/launcher) existe em FFX e FFX-2. Os painéis de
// ajuda (.sps2 em help/) são FFX-only: a árvore do ffx2 não tem a pasta help/.
export function entryKindsFor(version: GameVersionId): EntryKind[] {
  if (version === 'lastmiss') return ['events', 'objects'];
  if (version === 'ffx') return ['events', 'objects', 'macro', 'lockit', 'help'];
  return ['events', 'objects', 'macro', 'lockit'];
}

/**
 * Kinds de IMAGEM (.dds.phyre). Só ffx e ffx2: a Last Mission divide a
 * árvore do ffx2 mas não tem texturas próprias.
 *
 * Fica separado de entryKindsFor de propósito — esse é o conjunto de kinds
 * de TEXTO (escopo do Exportar/JSON, save-all e preload), e imagem não
 * participa de nada disso.
 */
export function imageKindsFor(version: GameVersionId): EntryKind[] {
  if (version === 'lastmiss') return [];
  return ['images'];
}

/** Todos os kinds da árvore da versão (texto + imagem), em ordem canônica. */
export function allKindsFor(version: GameVersionId): EntryKind[] {
  return [...entryKindsFor(version), ...imageKindsFor(version)];
}

/**
 * Fluxo novo (DTO): o backend envia o sumário canônico (id + key) e o DTO
 * estruturado sob demanda. Os artefatos de export/import são JSON + .strings
 * em mods/edits; não há mais árvore de diretórios nem .txt.
 */
export async function loadKindEntries(
  kind: EntryKind,
  version: GameVersionId
): Promise<EntryRow[]> {
  const summaries: services.EntrySummary[] = await ListTextEntries(
    kind,
    version as Parameters<typeof ListTextEntries>[1]
  );
  return (summaries ?? [])
    .map((s) => ({
      kind,
      id: s.id,
      key: s.key,
      label: resolveEntryLabel(kind, s.id),
    }))
    .sort((a, b) => a.label.localeCompare(b.label));
}

export async function loadEntry(
  kind: EntryKind,
  id: string,
  version: GameVersionId
): Promise<dto.FileEntry> {
  return GetTextEntry(
    kind,
    id,
    version as Parameters<typeof GetTextEntry>[2]
  );
}

// ---- Imagens (.dds.phyre) ----
// kind=images não é texto: não tem rows, não participa de export/import de
// texto. Cada chamada abaixo espelha um binding do App.

/** Carrega a textura e devolve a pré-visualização (data URL) + metadados. */
export async function loadImage(
  id: string,
  version: GameVersionId
): Promise<dto.ImageEntry> {
  return GetImageEntry(
    'images',
    id,
    version as Parameters<typeof GetImageEntry>[2]
  );
}

/** Reempacota um .dds sobre o container pristine e grava em mods/. */
export function importImage(
  id: string,
  ddsPath: string,
  version: GameVersionId
): Promise<void> {
  return ImportImage(
    'images',
    id,
    ddsPath,
    version as Parameters<typeof ImportImage>[3]
  );
}

/**
 * A textura ainda existe (data/ OU mods/)? Usada na revalidação da seleção
 * após reload: id apagado pelo delete NÃO deve ser re-carregado — o erro
 * "não encontrada em data/ nem em mods/" só faria confusão. Falha de
 * binding = não existe (não serve para exibição).
 */
export async function imageExists(
  id: string,
  version: GameVersionId
): Promise<boolean> {
  try {
    return await ImageExists('images', id, version);
  } catch {
    return false;
  }
}

/**
 * Reempacota o MESMO .dds na textura e nas cópias idênticas dela (payload
 * original igual — as réplicas da otimização do DVD). O backend valida o
 * lote antes de gravar: alvo fora do grupo é recusado inteiro.
 */
export function importImageGroup(
  id: string,
  ddsPath: string,
  targets: string[],
  version: GameVersionId
): Promise<dto.ImageImportResult> {
  return ImportImageGroup(
    'images',
    id,
    ddsPath,
    targets,
    version as Parameters<typeof ImportImageGroup>[4]
  );
}

/**
 * Descarta o índice de duplicatas em cache (rebuild na próxima abertura da
 * textura) — existe para edição EXTERNA em hex editor, que o cache não vê.
 */
export function refreshImageDuplicates(version: GameVersionId): Promise<void> {
  return RefreshImageDuplicates(
    'images',
    version as Parameters<typeof RefreshImageDuplicates>[1]
  );
}

/**
 * Cópias da textura SEM carregar a imagem (só o índice memoizado): é o que
 * o menu de contexto e os diálogos de extrair/replicar/deletar pedem para
 * agir sobre o NÓ CLICADO, que pode não ser a entry selecionada.
 */
export function imageDuplicates(
  id: string,
  version: GameVersionId
): Promise<dto.ImageDuplicates> {
  return ImageDuplicates(
    'images',
    id,
    version as Parameters<typeof ImageDuplicates>[2]
  );
}

/** Extrai .dds + .png desta textura e das cópias escolhidas (lote). */
export function extractImageGroup(
  id: string,
  targets: string[],
  version: GameVersionId
): Promise<dto.BatchResult> {
  return ExtractImageGroup(
    'images',
    id,
    targets,
    version as Parameters<typeof ExtractImageGroup>[3]
  );
}

/**
 * Reempacota a IMAGEM ABERTA em mods/ das cópias escolhidas (o "dupe"): a
 * fonte é o próprio conteúdo servido ao painel, sem diálogo de arquivo.
 */
export function replicateImage(
  id: string,
  targets: string[],
  version: GameVersionId
): Promise<dto.BatchResult> {
  return ReplicateImage(
    'images',
    id,
    targets,
    version as Parameters<typeof ReplicateImage>[3]
  );
}

/**
 * Apaga os containers escolhidos no escopo pedido e SEMPRE os artefatos
 * derivados (.dds/.png extraídos). Escopo vem explícito do diálogo.
 */
export function deleteImages(
  id: string,
  targets: string[],
  scope: 'data' | 'mods' | 'both',
  version: GameVersionId
): Promise<dto.BatchResult> {
  return DeleteImages(
    'images',
    id,
    targets,
    scope,
    version as Parameters<typeof DeleteImages>[4]
  );
}

/**
 * "Abrir até o arquivo": abre o explorador já com o arquivo da entrada
 * selecionado, resolvido mods-first (o arquivo que está valendo).
 */
export function revealEntry(
  kind: EntryKind,
  id: string,
  version: GameVersionId
): Promise<void> {
  return RevealEntryFile(
    kind,
    id,
    version as Parameters<typeof RevealEntryFile>[2]
  );
}

/** Salva .dds ou .png no caminho escolhido pelo usuário. */
export function saveImage(
  id: string,
  format: 'dds' | 'png',
  destPath: string,
  version: GameVersionId
): Promise<void> {
  return SaveImage(
    'images',
    id,
    format,
    destPath,
    version as Parameters<typeof SaveImage>[4]
  );
}

/** Seletor nativo para escolher um .dds a importar. "" = cancelado. */
export function selectImageFile(): Promise<string> {
  return SelectImageFile();
}

/** Diálogo "Salvar como" para .dds/.png. "" = cancelado. */
export function selectImageSavePath(
  format: 'dds' | 'png',
  suggestedName: string
): Promise<string> {
  return SelectImageSavePath(format, suggestedName);
}

/** Exporta a entrada em JSON + .strings (mods/edits). */
export async function exportEntry(
  kind: EntryKind,
  id: string,
  version: GameVersionId,
  langs: string[] = []
): Promise<string[]> {
  return ExportEntry(
    kind,
    id,
    version as Parameters<typeof ExportEntry>[2],
    langs
  );
}

/** Lê o artefato JSON padrão em mods/edits e aplica no binário. */
export async function importEntry(
  kind: EntryKind,
  id: string,
  version: GameVersionId
): Promise<string[]> {
  return ImportEntry(
    kind,
    id,
    version as Parameters<typeof ImportEntry>[2]
  );
}

/** Exporta o lote em JSON (ids vazio = tudo do kind; langs vazio = todos). */
export async function exportJSON(
  kind: EntryKind,
  version: GameVersionId,
  ids: string[],
  langs: string[] = []
): Promise<string[]> {
  return ExportJSON(kind, version as Parameters<typeof ExportJSON>[1], ids, langs);
}

/** Exporta o lote em .strings (mesmo contrato de exportJSON). */
export async function exportStrings(
  kind: EntryKind,
  version: GameVersionId,
  ids: string[],
  langs: string[] = []
): Promise<string[]> {
  return ExportStrings(kind, version as Parameters<typeof ExportStrings>[1], ids, langs);
}

/**
 * Resumo do destino do export para o toast: os artefatos saem em mods/edits
 * (todos os kinds), então mostra a contagem + diretório (ou o caminho único).
 */
export function exportDestination(paths: string[]): string {
  const list = paths ?? [];
  if (list.length === 0) return '';
  if (list.length === 1) return list[0];
  const dir = list[0].replace(/[\\/][^\\/]*$/, '');
  return `${list.length} arquivos em ${dir}`;
}

/** Seletor nativo de arquivo (.json/.strings). "" = cancelado. */
export function selectImportFile(): Promise<string> {
  return SelectImportFile();
}

/** Valida o arquivo e devolve o resumo do modal (sem aplicar nada). */
export async function previewImport(
  path: string,
  version: GameVersionId
): Promise<dto.ImportSummary> {
  return PreviewImport(path, version as Parameters<typeof PreviewImport>[1]);
}

/** Aplica o arquivo validado; devolve quantos textos 'us' mudaram. */
export async function importFile(
  path: string,
  version: GameVersionId
): Promise<number> {
  return ImportFile(path, version as Parameters<typeof ImportFile>[1]);
}
