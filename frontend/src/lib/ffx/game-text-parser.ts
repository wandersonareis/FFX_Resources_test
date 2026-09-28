import type { JSONContent } from '@tiptap/core';
import {
  COLOR_RESET_TAG,
  extractChipLabel,
  getColorTagName,
  getResolvedColorFromTag,
  isLockedTagInner,
} from './game-text-tags';
import type { ChipValidator } from './tag-catalog';

/**
 * Conversão entre o texto canônico do backend (tags {…}) e o HTML do Tiptap.
 *
 * Vira formatação rica: {CLR:X}/{COLOR:X} -> cor, {TEXT_ITALIC}/{TEXT_NORMAL}
 * -> itálico, \n / {TEXT_NEWLINE} -> quebra de parágrafo.
 * Todo o resto vira chip atômico <span data-game-tag="…"> com value = tag
 * original (serialização verbatim, byte-idêntica).
 */
export const gameTextParser = {
  /**
   * Texto canônico com tags -> HTML do Tiptap.
   *
   * `validate` (opcional) marca como inválido o chip da tag digitada que não
   * sobreviveria ao round-trip — a MESMA instância usada na conversão de tags
   * digitadas, senão o HTML divergiria do valor e o setContent perderia o
   * cursor.
   */
  parseGameTextToHTML(rawText: string, validate?: ChipValidator): string {
    if (!rawText) return '<p></p>';

    // Quebras: \n real ou {TEXT_NEWLINE} viram parágrafos.
    const paragraphs = rawText.split(/(?:\r?\n|\{TEXT_NEWLINE\})/gi);
    const htmlParagraphs: string[] = [];

    for (const para of paragraphs) {
      if (para.trim() === '') {
        htmlParagraphs.push('<p></p>');
        continue;
      }

      let currentColor: string | null = null;
      let italicActive = false;
      let resultHTML = '';
      let lastIndex = 0;

      const tagRegex = /\{([^{}]+)\}/g;
      let match: RegExpExecArray | null;

      while ((match = tagRegex.exec(para)) !== null) {
        const matchIndex = match.index;
        const inner = match[1].trim();

        const colorTag = this.classifyColorTag(inner);
        const fontTag = this.classifyFontTag(inner);

        if (!colorTag && !fontTag && !this.isChipTag(inner)) {
          continue;
        }

        if (matchIndex > lastIndex) {
          const segment = para.slice(lastIndex, matchIndex);
          if (segment) {
            resultHTML += this.wrapStyledText(segment, currentColor, italicActive);
          }
        }

        if (colorTag) {
          currentColor = colorTag.isReset ? null : colorTag.hex;
        } else if (fontTag) {
          italicActive = fontTag === 'italic';
        } else {
          // Chip atômico genérico: value = tag original, label derivado.
          const fullTag = `{${inner}}`;
          const label = extractChipLabel(inner);
          const locked = isLockedTagInner(inner);
          const invalid = !!validate && !validate(inner);
          // Ordem idêntica ao renderHTML do nó gameTag (data-game-tag, class,
          // data-locked, data-invalid): a comparação com getHTML() é literal
          // e uma ordem diferente dispararia setContent à toa.
          resultHTML += `<span data-game-tag="${this.escapeAttr(
            fullTag
          )}" class="game-tag-chip${locked ? ' locked' : ''}${
            invalid ? ' invalid' : ''
          }"${locked ? ' data-locked="true"' : ''}${
            invalid ? ' data-invalid="true"' : ''
          }>${this.escapeHtml(label)}</span>`;
        }

        lastIndex = tagRegex.lastIndex;
      }

      if (lastIndex < para.length) {
        const remaining = para.slice(lastIndex);
        if (remaining) {
          resultHTML += this.wrapStyledText(remaining, currentColor, italicActive);
        }
      }

      htmlParagraphs.push(`<p>${resultHTML || '<br>'}</p>`);
    }

    return htmlParagraphs.join('');
  },

  /**
   * Documento JSON do Tiptap -> texto canônico com tags.
   * Cores/itálico sincronizados por parágrafo; chips emitidos verbatim.
   * Parágrafos unidos com {TEXT_NEWLINE} (espelha WriteLinebreaksAsCommands).
   */
  serializeJSONToGameText(json: JSONContent): string {
    if (!json || !json.content) return '';

    const paragraphTexts: string[] = [];

    for (const block of json.content) {
      let blockText = '';
      let currentColorTag: string | null = null;
      let italicActive = false;

      if (block.content && block.content.length > 0) {
        for (const inlineNode of block.content) {
          if (inlineNode.type === 'gameTag') {
            const value = (inlineNode.attrs?.['value'] as string) || '';
            if (value) blockText += value;
          } else if (inlineNode.type === 'text') {
            const text = inlineNode.text || '';

            const colorMark = inlineNode.marks?.find(
              (m) => m.type === 'textStyle' && m.attrs?.['color']
            );
            const colorVal = colorMark?.attrs?.['color'] as string | undefined;
            const nodeColorTag = colorVal ? getColorTagName(colorVal) : null;

            const hasItalic = !!inlineNode.marks?.some((m) => m.type === 'italic');

            if (hasItalic !== italicActive) {
              blockText += hasItalic ? '{TEXT_ITALIC}' : '{TEXT_NORMAL}';
              italicActive = hasItalic;
            }

            if (nodeColorTag !== currentColorTag) {
              if (nodeColorTag) {
                blockText += `{CLR:${nodeColorTag}}`;
              } else if (currentColorTag) {
                blockText += `{CLR:${COLOR_RESET_TAG}}`;
              }
              currentColorTag = nodeColorTag;
            }

            blockText += text;
          } else if (inlineNode.type === 'hardBreak') {
            blockText += '{TEXT_NEWLINE}';
          }
        }

        if (currentColorTag) {
          blockText += `{CLR:${COLOR_RESET_TAG}}`;
          currentColorTag = null;
        }
        if (italicActive) {
          blockText += '{TEXT_NORMAL}';
          italicActive = false;
        }
      }

      paragraphTexts.push(blockText);
    }

    return paragraphTexts.join('{TEXT_NEWLINE}');
  },

  classifyColorTag(inner: string): { hex: string | null; isReset: boolean } | null {
    const upper = inner.trim().toUpperCase();
    if (upper.startsWith('CLR:')) {
      return getResolvedColorFromTag(upper.slice(4));
    }
    if (upper.startsWith('COLOR:')) {
      return getResolvedColorFromTag(upper.slice(6));
    }
    return null;
  },

  classifyFontTag(inner: string): 'italic' | 'normal' | null {
    const upper = inner.trim().toUpperCase();
    if (upper === 'TEXT_ITALIC') return 'italic';
    if (upper === 'TEXT_NORMAL') return 'normal';
    return null;
  },

  isChipTag(inner: string): boolean {
    const upper = inner.trim().toUpperCase();
    // TEXT_NEWLINE já foi consumido no split; se sobrar, trata como chip
    // para nunca perder conteúdo.
    return upper.length > 0;
  },

  wrapStyledText(text: string, color: string | null, italic: boolean): string {
    let escaped = this.escapeHtml(text);
    if (italic) escaped = `<em>${escaped}</em>`;
    if (color) escaped = `<span style="color: ${color}">${escaped}</span>`;
    return escaped;
  },

  /**
   * Escapa texto de nó (mesmo algoritmo da serialização HTML): só `&`, `<`,
   * `>` e nbsp. Escapar aspas (como `&quot;`/`&#039;`) tornaria o HTML
   * diferente do getHTML() — a comparação é literal e o setContent só roda
   * quando o documento realmente divergiu.
   */
  escapeHtml(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/\u00A0/g, '&nbsp;');
  },

  /**
   * Escapa valor de atributo (idem serialização HTML): `&`, `"`, nbsp e CR.
   */
  escapeAttr(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/"/g, '&quot;')
      .replace(/\u00A0/g, '&nbsp;')
      .replace(/\r/g, '&#13;');
  },
};

