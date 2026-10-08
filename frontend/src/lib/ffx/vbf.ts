import {
  ExportVbfSelection,
  ExtractVbfSelection,
  GetVbfImageEntry,
  GetVbfTextEntry,
  ListVbfDir,
  ListVbfMacroChunks,
  ListVbfRoots,
  PreviewVbfExtraction,
  SelectVbfExtractDir,
} from '@/wailsjs/go/main/App';
import { dto } from '@/wailsjs/go/models';

/**
 * Navegador de container .vbf (SOMENTE LEITURA).
 *
 * O backend nunca lê o container inteiro e nunca escreve nele: a árvore sai
 * do índice (cabeçalho) e só o arquivo clicado é decodificado, em memória,
 * dentro de um escopo que some no fim da chamada. Aqui fica a camada de
 * acesso + um cache LRU para reabrir o que já foi decodificado sem pagar o
 * decode de novo (trocar de nó na árvore é a interação mais frequente).
 */

/** Raiz descoberta perto do executável do jogo. */
export type VbfRoot = dto.VbfRoot;

/** Filho de um diretório do .vbf — ou chunk virtual do macrodic. */
export type VbfNode = dto.VbfNode;

/**
 * Cache LRU por chave: mantém a ordem de acesso e derruba o mais antigo.
 * Só guarda o que já foi pedido — nada aqui é pré-extraído.
 */
class Lru<T> {
  private readonly map = new Map<string, T>();

  constructor(private readonly capacity: number) {}

  get(key: string): T | undefined {
    const value = this.map.get(key);
    if (value === undefined) return undefined;
    // Toca para mover para o fim (mais recente).
    this.map.delete(key);
    this.map.set(key, value);
    return value;
  }

  set(key: string, value: T): void {
    this.map.delete(key);
    this.map.set(key, value);
    while (this.map.size > this.capacity) {
      const oldest = this.map.keys().next();
      if (oldest.done) break;
      this.map.delete(oldest.value);
    }
  }

  clear(): void {
    this.map.clear();
  }
}

const dirCache = new Lru<VbfNode[]>(64);
const entryCache = new Lru<dto.FileEntry>(24);
const imageCache = new Lru<dto.ImageEntry>(12);
let rootsPromise: Promise<VbfRoot[]> | null = null;

/**
 * Raízes .vbf (uma promessa compartilhada: todas as abas veem a mesma
 * lista e o backend só varre a pasta do executável uma vez).
 */
export function loadVbfRoots(): Promise<VbfRoot[]> {
  rootsPromise ??= ListVbfRoots().catch((error) => {
    rootsPromise = null;
    throw error;
  });
  return rootsPromise;
}

/** Filhos imediatos de um diretório do .vbf (diretórios primeiro). */
export function loadVbfDir(root: string, dir: string): Promise<VbfNode[]> {
  const key = `${root}|${dir.toLowerCase()}`;
  const cached = dirCache.get(key);
  if (cached) return Promise.resolve(cached);
  return ListVbfDir(root, dir).then((nodes) => {
    dirCache.set(key, nodes ?? []);
    return nodes ?? [];
  });
}

/**
 * Filhos virtuais de macrodic.dcp (chunk_00, chunk_01, …): o dicionário é
 * UM arquivo por localização e o app o trata como grupo.
 */
export function loadVbfMacroChunks(
  root: string,
  macroPath: string
): Promise<VbfNode[]> {
  const key = `macro|${root}|${macroPath.toLowerCase()}`;
  const cached = dirCache.get(key);
  if (cached) return Promise.resolve(cached);
  return ListVbfMacroChunks(root, macroPath).then((nodes) => {
    dirCache.set(key, nodes ?? []);
    return nodes ?? [];
  });
}

/** Decodifica um arquivo de TEXTO do .vbf e devolve a tabela. */
export function loadVbfEntry(
  root: string,
  path: string,
  id: string
): Promise<dto.FileEntry> {
  const key = `${root}|${path.toLowerCase()}|${id}`;
  const cached = entryCache.get(key);
  if (cached) return Promise.resolve(cached);
  return GetVbfTextEntry(root, path, id).then((entry) => {
    entryCache.set(key, entry);
    return entry;
  });
}

/** Decodifica uma textura (.dds.phyre) do .vbf (data URL + metadados). */
export function loadVbfImage(
  root: string,
  path: string
): Promise<dto.ImageEntry> {
  const key = `${root}|${path.toLowerCase()}`;
  const cached = imageCache.get(key);
  if (cached) return Promise.resolve(cached);
  return GetVbfImageEntry(root, path).then((image) => {
    imageCache.set(key, image);
    return image;
  });
}

/**
 * Descarta tudo: o executável do jogo mudou (outra pasta de containers) ou
 * o usuário pediu reload da árvore — os caminhos antigos não valem mais.
 */
export function invalidateVbfCache(): void {
  dirCache.clear();
  entryCache.clear();
  imageCache.clear();
  rootsPromise = null;
}

// ---- extração/export da SELEÇÃO da árvore do .vbf --------------------------
//
// A seleção é uma lista de caminhos internos (arquivos e diretórios);
// diretório = toda a subárvore (expandida contra o índice no backend).

/** Resumo antes de extrair: arquivos, bytes e quantos já existem no destino. */
export function previewVbfExtraction(
  root: string,
  paths: string[],
  destRoot: string
): Promise<dto.VbfExtractPreview> {
  return PreviewVbfExtraction(root, paths, destRoot);
}

/** Extrai a seleção preservando a estrutura de caminhos em destRoot. */
export function extractVbfSelection(
  root: string,
  paths: string[],
  destRoot: string
): Promise<dto.BatchResult> {
  return ExtractVbfSelection(root, paths, destRoot);
}

/** Seletor nativo de pasta do "Extrair para…" (default data/). "" = cancelou. */
export function selectVbfExtractDir(): Promise<string> {
  return SelectVbfExtractDir();
}

/**
 * Exporta os kinds de TEXTO da seleção em JSON ou .strings — mesmo lugar
 * (mods/edits) e convenção do export de data/.
 */
export function exportVbfSelection(
  root: string,
  format: 'json' | 'strings',
  paths: string[],
  langs: string[] = []
): Promise<string[]> {
  return ExportVbfSelection(root, format, paths, langs);
}
