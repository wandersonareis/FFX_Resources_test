// Texto 'us' em branco — espelho do isBlankText do backend
// (backend/services/import.service.go): vazio, só espaço ou o "-" que é o
// próprio texto do jogo. Export e import partem deste predicado; a tabela
// da UI mostra só o que o artefato exportado mostraria.

/**
 * Texto não traduzível no idioma padrão: vazio, só espaço, ou o hífen
 * "-" (placeholder do jogo, ~metade do data/ pristine).
 */
export function isBlankText(t: string | undefined | null): boolean {
  const s = t?.trim() ?? '';
  return s === '' || s === '-';
}

/**
 * Row da view sem 'us' utilizável — fica fora da TABELA e da contagem de
 * progresso do badge, sempre no idioma padrão (o backend exporta só o que
 * tem 'us': se o 'us' não existe, os outros idiomas não interessam).
 */
export function isBlankSourceRow(text: string | undefined | null): boolean {
  return isBlankText(text);
}
