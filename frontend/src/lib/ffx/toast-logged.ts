import { toast } from 'sonner';
import { WriteLog } from '@/wailsjs/go/main/App';

type SonnerMessage = Parameters<typeof toast>[0];
type SonnerData = Parameters<typeof toast>[1];

/**
 * Copia para o log em arquivo (binding WriteLog → loggingService.FromFrontend
 * → logs/ffx-*.log em JSON) tudo que aparece como toast. As linhas já saem
 * com "source":"frontend" e aqui o field kind=toast completa a marcação — no
 * arquivo fica EXPLÍCITO que é um toast do frontend, não um fluxo do backend.
 *
 * Regra central de duração: toast de ERRO fica 20 s em tela. (Antes: duração
 * Infinity sem botão de fechar habilitado — toast eterno.) Os demais seguem
 * o padrão do sonner (4 s).
 */
function logToBackend(level: string, message: SonnerMessage): void {
  WriteLog(level, typeof message === 'string' ? message : '[conteúdo não textual]', {
    kind: 'toast',
  }).catch(() => {
    // Log é best-effort: falha de log nunca dispara outro toast nem bloqueia a UI.
  });
}

/** toast com cópia integral para o log do backend. */
function neutralToast(message: SonnerMessage, data?: SonnerData) {
  logToBackend('info', message);
  return toast(message, data);
}

export const loggedToast = Object.assign(neutralToast, {
  error: (message: SonnerMessage, data?: SonnerData) => {
    logToBackend('error', message);
    return toast.error(message, { duration: 20_000, ...data });
  },
  info: (message: SonnerMessage, data?: SonnerData) => {
    logToBackend('info', message);
    return toast.info(message, data);
  },
  success: (message: SonnerMessage, data?: SonnerData) => {
    logToBackend('info', message);
    return toast.success(message, data);
  },
  warning: (message: SonnerMessage, data?: SonnerData) => {
    logToBackend('warn', message);
    return toast.warning(message, data);
  },
  loading: (message: SonnerMessage, data?: SonnerData) => {
    logToBackend('info', message);
    return toast.loading(message, data);
  },
  message: (message: SonnerMessage, data?: SonnerData) => {
    logToBackend('info', message);
    return toast.message(message, data);
  },
});
