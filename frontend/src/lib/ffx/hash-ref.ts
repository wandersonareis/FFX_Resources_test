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
 * Linha é referência de dedup de `text` (repetição NÃO traduzida de um
 * original definido em outro ponto) no idioma dado? As refs chegam do
 * backend no DTO dedupado e ficam fora da tabela — o tradutor traduz cada
 * original uma vez; as cópias herdam a tradução da def no salvar (o
 * ponteiro é endereçado pelo ORIGINAL de data/, não pelo texto traduzido).
 * Texto divergente da def é carga real de revisão e permanece literal.
 */
export function isRefText(
  text: string | undefined | null,
  hash: string | undefined | null
): boolean {
  if (!text || !hash) return false;
  return text.startsWith('$') && text.slice(1) === hash;
}
