import { Node } from '@tiptap/core';

export interface GameTagAttrs {
  /** Tag original completa, ex: {MCR:s06:l3F:"Yevon"}. Serializada verbatim. */
  value: string;
  /** Rótulo exibido no chip, ex: Yevon. */
  label: string;
  /** CMD, HEX, UNK ou VAR: valores calculados no jogo, nunca editáveis. */
  locked: boolean;
}

/**
 * Nó inline atômico para tags de controle (valor humano do terceiro campo
 * ignorado na serialização — só `value` é emitido, verbatim).
 * Renderização estática via renderHTML (sem NodeView Angular).
 */
export const GameTagExtension = Node.create<GameTagAttrs>({
  name: 'gameTag',
  group: 'inline',
  inline: true,
  atom: true,
  selectable: true,
  draggable: false,

  addAttributes() {
    return {
      value: {
        default: '',
        parseHTML: (element) => element.getAttribute('data-game-tag') || '',
      },
      label: {
        default: '',
        parseHTML: (element) => element.textContent || '',
      },
      locked: {
        default: false,
        parseHTML: (element) => element.hasAttribute('data-locked'),
      },
    };
  },

  parseHTML() {
    return [{ tag: 'span[data-game-tag]' }];
  },

  renderHTML({ node }) {
    const value = node?.attrs['value'] ?? '';
    const label = node?.attrs['label'] ?? value;
    const locked = !!node?.attrs['locked'];
    const attrs: Record<string, string> = {
      'data-game-tag': value,
      class: locked ? 'game-tag-chip locked' : 'game-tag-chip',
    };
    if (locked) attrs['data-locked'] = 'true';
    return ['span', attrs, label];
  },
});
