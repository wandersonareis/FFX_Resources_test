import type { JSONContent } from '@tiptap/core';
import {
  COLOR_RESET_TAG,
  extractChipLabel,
  getColorTagName,
  getResolvedColorFromTag,
  isLockedTagInner,
} from './game-text-tags';

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
   */
  parseGameTextToHTML(rawText: string): string {
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
          resultHTML += `<span data-game-tag="${this.escapeAttr(
            fullTag
          )}"${locked ? ' data-locked="true"' : ''} class="game-tag-chip${
            locked ? ' locked' : ''
          }">${this.escapeHtml(label)}</span>`;
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

  escapeHtml(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  },

  escapeAttr(text: string): string {
    return this.escapeHtml(text);
  },
};

