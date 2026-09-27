// Resolução de refs de dedup ($hash) — espelho da regra do backend
// (unmarshalCollection / helpRefBare): o texto é referência quando começa
// com "$" E o resto casa com o hash PRÓPRIO da row. Prefixo sozinho não
// basta: um texto literal que comece com "$" não pode virar linha oculta.

/**
 * Kinds cujo view sai com dedup global (refs "$hash" ocultas na tabela):
 * os formatos de extração única/agregada — help (arquivo único com os 6
 * painéis), events (eventos gêmeos com repetição maciça), macro (dicionário
 * único em um artefato) e objects (repetição massiva intra-arquivo; refs
 * não propagam entre objetos).
 */
export const DEDUP_VIEW_KINDS: ReadonlySet<string> = new Set([
  'help',
  'events',
  'macro',
  'objects',
]);

/**
 * Linha é referência de dedup de `text` (repetição idêntica de outro
 * ponto/painel) no idioma dado? As refs chegam do backend no DTO dedupado
 * e ficam fora da tabela — o tradutor vê cada texto uma vez; o resolve/
 * propaga do texto editado é do backend no salvar.
 */
export function isRefText(
  text: string | undefined | null,
  hash: string | undefined | null
): boolean {
  if (!text || !hash) return false;
  return text.startsWith('$') && text.slice(1) === hash;
}
