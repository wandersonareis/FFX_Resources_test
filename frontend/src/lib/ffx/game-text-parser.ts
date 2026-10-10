import type { JSONContent } from '@tiptap/core';
import {
  COLOR_RESET_TAG,
  buttonSpriteClasses,
  extractChipLabel,
  getColorTagName,
  getResolvedColorFromTag,
  isLockedTagInner,
} from './game-text-tags';
import type { ChipValidator } from './tag-catalog';

/** Destaque opcional de um intervalo do texto (usado pelo snippet da busca). */
export interface GameTextMarkOptions {
  /** Início (inclusivo) em CODE POINTS do texto de entrada. */
  markStart?: number;
  /** Fim (exclusivo) em CODE POINTS do texto de entrada. */
  markEnd?: number;
}

/** Offset em code points -> índice UTF-16 (clamp nas pontas). */
function toUtf16Index(text: string, codePoints: number): number {
  let count = 0;
  for (let i = 0; i < text.length; ) {
    if (count >= codePoints) return i;
    const code = text.codePointAt(i) ?? 0;
    i += code > 0xffff ? 2 : 1;
    count += 1;
  }
  return text.length;
}

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
   *
   * `options` (opcional) destaca um intervalo do texto em <mark> — usado
   * pelo snippet da busca. Sem ele, a saída é byte-idêntica à de sempre.
   */
  parseGameTextToHTML(
    rawText: string,
    validate?: ChipValidator,
    options?: GameTextMarkOptions
  ): string {
    if (!rawText) return '<p></p>';

    // Destaque (busca): markStart/markEnd em CODE POINTS do texto cru,
    // resolvidos uma única vez para UTF-16 — a varredura trabalha em
    // offsets absolutos e nunca corta uma tag no meio.
    const mark = this.resolveMarkRange(rawText, options);

    // Quebras: \n real ou {TEXT_NEWLINE} viram parágrafos. O split leva
    // grupo de captura (separadores intercalados no resultado), então o
    // cursor absoluto avança exatamente pelo que foi consumido.
    const segments = rawText.split(/(\r?\n|\{TEXT_NEWLINE\})/gi);
    const htmlParagraphs: string[] = [];
    let cursor = 0;

    for (let s = 0; s < segments.length; s += 2) {
      const para = segments[s];
      const paraStart = cursor;
      cursor +=
        para.length + (s + 1 < segments.length ? segments[s + 1].length : 0);

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
            resultHTML += this.wrapSegment(
              segment,
              paraStart + lastIndex,
              currentColor,
              italicActive,
              mark
            );
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
          // BUTTON com glifo(s) mapeado(s) e sem marcação: os ícones vão
          // DIRETO no texto (sem chip) — igual aos botões do jogo. Chip
          // textual fica para os casos com texto (locked/invalid/dummy).
          const btnClasses = buttonSpriteClasses(inner);
          let chipHTML: string;
          if (btnClasses.length > 0 && !locked && !invalid) {
            // O wrapper externo (data-game-tag) é o ponto de re-parse do
            // node gameTag; os ícones ficam ANINHADOS dentro dele — um node
            // atômico do ProseMirror só aceita UM elemento raiz no
            // renderHTML, e irmãos no mesmo nível virariam filhos do
            // primeiro tile (tile tem tamanho fixo: sobrepõem). Estrutura
            // byte-idêntica à do renderHTML do gameTag.
            chipHTML =
              `<span data-game-tag="${this.escapeAttr(fullTag)}">` +
              btnClasses
                .map(
                  (cls) =>
                    `<span class="gb-sprite ${cls}" title="${this.escapeAttr(
                      label
                    )}"></span>`
                )
                .join('') +
              '</span>';
          } else {
            chipHTML = `<span data-game-tag="${this.escapeAttr(
              fullTag
            )}" class="game-tag-chip${locked ? ' locked' : ''}${
              invalid ? ' invalid' : ''
            }"${locked ? ' data-locked="true"' : ''}${
              invalid ? ' data-invalid="true"' : ''
            }>${this.escapeHtml(label)}</span>`;
          }
          // Destaque que intersecta a TAG inteira: o chip ganha o <mark>
          // (não há texto fora da braces para envolver — senão o hit
          // ficaria invisível).
          const tagStart = paraStart + matchIndex;
          const tagEnd = paraStart + tagRegex.lastIndex;
          resultHTML +=
            mark && mark.start < tagEnd && mark.end > tagStart
              ? `<mark>${chipHTML}</mark>`
              : chipHTML;
        }

        lastIndex = tagRegex.lastIndex;
      }

      if (lastIndex < para.length) {
        const remaining = para.slice(lastIndex);
        if (remaining) {
          resultHTML += this.wrapSegment(
            remaining,
            paraStart + lastIndex,
            currentColor,
            italicActive,
            mark
          );
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

  wrapSegment(
    text: string,
    absoluteStart: number,
    color: string | null,
    italic: boolean,
    mark: { start: number; end: number } | null
  ): string {
    const absoluteEnd = absoluteStart + text.length;
    if (!mark || mark.end <= absoluteStart || mark.start >= absoluteEnd) {
      return this.wrapStyledText(text, color, italic);
    }
    // O corte é ANTES do escape (escapar muda o comprimento); as 3 partes
    // são escapadas individualmente e o <mark> entra no meio, POR DENTRO
    // de itálico/cor — mesma aparência da tabela.
    const from = Math.max(mark.start, absoluteStart) - absoluteStart;
    const to = Math.min(mark.end, absoluteEnd) - absoluteStart;
    const escaped =
      this.escapeHtml(text.slice(0, from)) +
      `<mark>${this.escapeHtml(text.slice(from, to))}</mark>` +
      this.escapeHtml(text.slice(to));
    return this.wrapStyledTextHtml(escaped, color, italic);
  },

  wrapStyledText(text: string, color: string | null, italic: boolean): string {
    return this.wrapStyledTextHtml(this.escapeHtml(text), color, italic);
  },

  /** Itálico/cor sobre HTML JÁ escapado (pode conter <mark> interno). */
  wrapStyledTextHtml(
    html: string,
    color: string | null,
    italic: boolean
  ): string {
    if (italic) html = `<em>${html}</em>`;
    if (color) html = `<span style="color: ${color}">${html}</span>`;
    return html;
  },

  /** markStart/markEnd em code points -> intervalo UTF-16 (null = sem destaque). */
  resolveMarkRange(
    rawText: string,
    options?: GameTextMarkOptions
  ): { start: number; end: number } | null {
    if (options?.markStart == null || options?.markEnd == null) return null;
    const start = toUtf16Index(rawText, Math.max(0, options.markStart));
    const end = toUtf16Index(
      rawText,
      Math.max(options.markEnd, options.markStart)
    );
    return end > start ? { start, end } : null;
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

