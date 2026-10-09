const ACTIVE_TAB_KEY = 'ffx-resources.active-tab';

/**
 * Índice persistido pode vir de uma build anterior (lista de versões menor).
 * Fora da faixa o React nunca casa nenhuma aba — devolve 0, a primeira.
 */
export function normalizeTabIndex(index: number, count: number): number {
  return Number.isInteger(index) && index >= 0 && index < count ? index : 0;
}

/** Persiste a última aba selecionada no localStorage. */
export const tabStorage = {
  load(count: number): number {
    try {
      const raw = localStorage.getItem(ACTIVE_TAB_KEY);
      const index = raw === null ? 0 : Number.parseInt(raw, 10);
      return normalizeTabIndex(index, count);
    } catch {
      return 0;
    }
  },

  save(index: number): void {
    try {
      localStorage.setItem(ACTIVE_TAB_KEY, String(index));
    } catch {
      // localStorage indisponível: ignora.
    }
  },
};
