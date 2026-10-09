import { formatForDisplay } from '@tanstack/react-hotkeys';

/**
 * Catálogo dos ATALHOS DE AÇÃO: qual tecla, o que faz e onde vive o controle.
 * Fonte única — o registro (`useHotkey`/`useHotkeys`) e o texto de UI
 * (tooltip/`title`) leem daqui, então a tecla nunca é escrita em dois lugares.
 *
 * `as const` é obrigatório: `useHotkey` exige o tipo literal `Hotkey`
 * (`string` não compila) e um `Mod+${i}` dinâmico também não serviria.
 *
 * Fica de fora a navegação contextual (setas/Enter da árvore e da tabela,
 * ←/→ do diálogo): não aparece em texto nenhum e só faz sentido dentro do
 * contexto de cada componente.
 */
export const SHORTCUTS = {
  /** app-shell → grava os rascunhos no binário (botão "Salvar" da aba). */
  save: { keys: 'Mod+S', label: 'Salvar no binário' },
  /** app-shell → abre o ConfigDialog (engrenagem do header). */
  config: { keys: 'Mod+,', label: 'Abrir Configurações' },
  /** app-shell → uma tecla por aba de GAME_VERSIONS (1 = FFX, 2 = EC, ...). */
  versionTabs: {
    keys: ['Mod+1', 'Mod+2', 'Mod+3', 'Mod+4'],
    label: 'Trocar de versão',
  },
  /** image-panel → espelha a pré-visualização (só na tela). */
  imageFlip: { keys: 'Mod+Alt+V', label: 'Espelhar imagem' },
  /** image-panel → ciclo de zoom da pré-visualização. */
  zoomIn: { keys: 'Mod+=', label: 'Aumentar o zoom' },
  zoomOut: { keys: 'Mod+-', label: 'Diminuir o zoom' },
  zoomReset: { keys: 'Mod+0', label: 'Voltar ao ajuste' },
} as const;

/** Tecla formatada para a plataforma: `Ctrl+S` no Windows, `⌘S` no macOS. */
export function hotkeyLabel(keys: string): string {
  return formatForDisplay(keys);
}

/** Texto pronto para tooltip/`title`: rótulo + tecla formatada. */
export function shortcutTip(label: string, keys: string): string {
  return `${label} · ${hotkeyLabel(keys)}`;
}
