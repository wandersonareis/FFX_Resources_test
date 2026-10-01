import { createStore, useSelector } from '@tanstack/react-store';

/**
 * Avisos do backend para a barra de status (evento "StatusWarning"): erros
 * de desalinhamento entre original e tradução publicados pelo serviço.
 *
 * Comportamento:
 *  - dedupe por ID estável (reemissão substitui o aviso);
 *  - dismiss ao abrir o alert e fechar: o ID descartado sai da barra e é
 *    ignorado por reemissões nesta sessão (limpa no reload do app).
 */
export interface StatusBarWarning {
  id: string;
  kind: string;
  entryId: string;
  version: string;
  severity: string;
  message: string;
  details?: Record<string, unknown>;
}

interface StatusBarState {
  warnings: StatusBarWarning[];
}

export const statusBarStore = createStore<StatusBarState>({ warnings: [] });

/** Descartados na sessão: reemissão do mesmo ID não volta à barra. */
const dismissedIds = new Set<string>();

export function pushStatusBarWarning(w: StatusBarWarning): void {
  if (!w?.id || dismissedIds.has(w.id)) return;
  statusBarStore.setState((prev) => {
    const idx = prev.warnings.findIndex((x) => x.id === w.id);
    if (idx >= 0) {
      const warnings = [...prev.warnings];
      warnings[idx] = w;
      return { warnings };
    }
    return { warnings: [...prev.warnings, w] };
  });
}

/** Descarta o aviso: sai da rotação e não volta por reemissão na sessão. */
export function dismissStatusBarWarning(id: string): void {
  if (id) dismissedIds.add(id);
  statusBarStore.setState((prev) => ({
    warnings: prev.warnings.filter((w) => w.id !== id),
  }));
}

export function useStatusBarWarnings(): StatusBarWarning[] {
  return useSelector(statusBarStore, (s) => s.warnings);
}
