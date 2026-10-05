const ACTIVE_TAB_KEY = 'ffx-resources.active-tab';

/** Persiste a última aba selecionada no localStorage. */
export const tabStorage = {
  load(): number {
    try {
      const raw = localStorage.getItem(ACTIVE_TAB_KEY);
      const index = raw === null ? 0 : Number.parseInt(raw, 10);
      return Number.isInteger(index) && index >= 0 ? index : 0;
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
