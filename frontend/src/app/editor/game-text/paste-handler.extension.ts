import { Extension } from '@tiptap/core';
import { Plugin } from '@tiptap/pm/state';
import { GameTextParserService } from './game-text-parser.service';

/**
 * Converte texto colado com tags canônicas {…} para o documento,
 * em vez do parse padrão do Tiptap.
 */
export function createPasteHandlerExtension(parser: GameTextParserService) {
  return Extension.create({
    name: 'pasteHandler',

    addProseMirrorPlugins() {
      const editor = this.editor;

      return [
        new Plugin({
          props: {
            handlePaste: (view, event) => {
              const clipboardData = event.clipboardData;
              if (!clipboardData) return false;

              const pastedText = clipboardData.getData('text/plain');
              if (!pastedText) return false;

              event.preventDefault();
              const convertedHtml = parser.parseGameTextToHTML(pastedText);
              editor.commands.insertContent(convertedHtml);
              return true;
            },
          },
        }),
      ];
    },
  });
}
