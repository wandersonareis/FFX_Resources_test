'use client';

import { useEffect, useRef, useState } from 'react';
import { Editor } from '@tiptap/core';
import StarterKit from '@tiptap/starter-kit';
import { TextStyle } from '@tiptap/extension-text-style';
import { Color } from '@tiptap/extension-color';
import { ChevronDown, Code, DoorOpen, Eraser, Gamepad2, Italic, Navigation, Palette, RotateCcw, Undo2, Redo2 } from 'lucide-react';
import { gameTextParser } from '@/lib/ffx/game-text-parser';
import { KNOWN_COLORS, extractChipLabel, isLockedTagInner } from '@/lib/ffx/game-text-tags';
import { GameTagExtension } from './game-tag.extension';
import { LockedTagsExtension } from './locked-tags.extension';
import { createPasteHandlerExtension } from './paste-handler.extension';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Separator } from '@/components/ui/separator';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip';

interface ControlTemplate {
  label: string;
  value: string;
}

const BUTTON_TEMPLATES: ControlTemplate[] = [
  { label: 'TRIANGLE', value: '{BUTTON:30:TRIANGLE}' },
  { label: 'X', value: '{BUTTON:31:X}' },
  { label: 'CIRCLE', value: '{BUTTON:32:CIRCLE}' },
  { label: 'SQUARE', value: '{BUTTON:33:SQUARE}' },
  { label: 'L1', value: '{BUTTON:34:L1}' },
  { label: 'R1', value: '{BUTTON:35:R1}' },
  { label: 'L2', value: '{BUTTON:36:L2}' },
  { label: 'R2', value: '{BUTTON:37:R2}' },
  { label: 'START', value: '{BUTTON:38:START}' },
  { label: 'SELECT', value: '{BUTTON:39:SELECT}' },
];

const DPAD_TEMPLATES: ControlTemplate[] = [
  { label: 'Direcional', value: '{BUTTON:40:Direcional}' },
  { label: 'Cima', value: '{BUTTON:41:Direcional UP}' },
  { label: 'Direita', value: '{BUTTON:42:Direcional RIGHT}' },
  { label: 'Baixo', value: '{BUTTON:44:Direcional DOWN}' },
  { label: 'Esquerda', value: '{BUTTON:48:Direcional LEFT}' },
  { label: 'Todos', value: '{BUTTON:4F:Direcional All}' },
];

const ICON_TEMPLATES: ControlTemplate[] = [
  { label: 'Red Gate', value: '{ICON:80:Red Gate}' },
  { label: 'Green Gate', value: '{ICON:81:Green Gate}' },
  { label: 'Yellow Gate', value: '{ICON:82:Yellow Gate}' },
  { label: 'Blue Gate', value: '{ICON:83:Blue Gate}' },
];

const COMMAND_TEMPLATES: ControlTemplate[] = [
  { label: 'PAUSE', value: '{PAUSE}' },
  { label: 'BREAK', value: '{BREAK}' },
  { label: 'Nova página', value: '{TEXT_NEWLINE}' },
  { label: 'CHOICE:00', value: '{CHOICE:00}' },
  { label: 'CHOICE-END', value: '{CHOICE-END}' },
];

export interface GameTextEditorProps {
  /** Texto canônico com tags (ex: linha de TextRow.text[lang]). */
  value: string;
  /** Emite o texto canônico serializado a cada edição. */
  onValueChange: (value: string) => void;
}

export function GameTextEditor({ value, onValueChange }: GameTextEditorProps) {
  const hostRef = useRef<HTMLDivElement | null>(null);
  const suppressUpdate = useRef(false);
  const onValueChangeRef = useRef(onValueChange);
  const [editor, setEditor] = useState<Editor | null>(null);
  const [, setTick] = useState(0);

  useEffect(() => {
    onValueChangeRef.current = onValueChange;
  }, [onValueChange]);

  useEffect(() => {
    const ed = new Editor({
      element: hostRef.current ?? undefined,
      extensions: [
        StarterKit.configure({
          heading: false,
          bulletList: false,
          orderedList: false,
          listItem: false,
          blockquote: false,
          codeBlock: false,
          horizontalRule: false,
          bold: false,
          strike: false,
        }),
        TextStyle,
        Color,
        GameTagExtension,
        LockedTagsExtension,
        createPasteHandlerExtension(gameTextParser),
      ],
      content: gameTextParser.parseGameTextToHTML(value),
      onUpdate: ({ editor: updated }) => {
        suppressUpdate.current = true;
        try {
          onValueChangeRef.current(
            gameTextParser.serializeJSONToGameText(updated.getJSON())
          );
        } finally {
          queueMicrotask(() => (suppressUpdate.current = false));
        }
      },
      onTransaction: () => setTick((t) => t + 1),
    });
    setEditor(ed);
    return () => {
      ed.destroy();
      setEditor(null);
    };
    // Cria o editor uma única vez.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!editor || suppressUpdate.current) return;
    const html = gameTextParser.parseGameTextToHTML(value);
    if (editor.getHTML() !== html) {
      editor.commands.setContent(html, { emitUpdate: false });
    }
  }, [editor, value]);

  const insertControlTag = (template: ControlTemplate) => {
    if (!editor) return;
    const inner = template.value.slice(1, -1);
    editor
      .chain()
      .focus()
      .insertContent({
        type: 'gameTag',
        attrs: {
          value: template.value,
          label: extractChipLabel(inner),
          locked: isLockedTagInner(inner),
        },
      })
      .run();
  };

  const colors = Object.values(KNOWN_COLORS).filter((c) => c.name !== 'WHITE');

  return (
    <div className="game-text-editor">
      <div className="editor-toolbar" role="toolbar" aria-label="Formatação do texto do jogo">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              className="toolbar-toggle"
              data-active={editor?.isActive('italic') ?? false}
              onClick={() => editor?.chain().focus().toggleItalic().run()}
              aria-label="Itálico"
            >
              <Italic size={20} />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Itálico ({'{TEXT_ITALIC}'})</TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => editor?.chain().focus().unsetItalic().run()}
            >
              <Eraser size={20} />
              Normal
            </Button>
          </TooltipTrigger>
          <TooltipContent>Texto normal ({'{TEXT_NORMAL}'})</TooltipContent>
        </Tooltip>

        <Separator orientation="vertical" className="toolbar-separator" />

        <TemplateDropdown
          label="Botões"
          icon={<Gamepad2 size={20} />}
          templates={BUTTON_TEMPLATES}
          tooltip="Inserir botão ({BUTTON:…})"
          onSelect={insertControlTag}
        />
        <TemplateDropdown
          label="Direcionais"
          icon={<Navigation size={20} />}
          templates={DPAD_TEMPLATES}
          tooltip="Inserir direcional ({BUTTON:…})"
          onSelect={insertControlTag}
        />
        <TemplateDropdown
          label="Ícones"
          icon={<DoorOpen size={20} />}
          templates={ICON_TEMPLATES}
          tooltip="Inserir ícone ({ICON:…})"
          onSelect={insertControlTag}
        />
        <TemplateDropdown
          label="Comandos"
          icon={<Code size={20} />}
          templates={COMMAND_TEMPLATES}
          tooltip="Inserir comando"
          onSelect={insertControlTag}
        />

        <Separator orientation="vertical" className="toolbar-separator" />

        <span className="color-group" role="group" aria-label="Cores ({CLR:…})">
          <Palette size={20} aria-label="Cores ({CLR:…}, {CLR:WHITE} reseta)" />
          {colors.map((c) => (
            <Tooltip key={c.name}>
              <TooltipTrigger asChild>
                <button
                  type="button"
                  className={`color-swatch${editor?.isActive('textStyle', { color: c.hex }) ? ' active' : ''}`}
                  style={{ backgroundColor: c.hex }}
                  aria-label={`Cor ${c.label}`}
                  onClick={() => editor?.chain().focus().setColor(c.hex).run()}
                />
              </TooltipTrigger>
              <TooltipContent>{`{CLR:${c.name}}`}</TooltipContent>
            </Tooltip>
          ))}
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                aria-label="Resetar cor"
                onClick={() => editor?.chain().focus().unsetColor().run()}
              >
                <RotateCcw size={20} />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Resetar cor ({'{CLR:WHITE}'})</TooltipContent>
          </Tooltip>
        </span>

        <Separator orientation="vertical" className="toolbar-separator" />

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              disabled={!editor?.can().undo()}
              aria-label="Desfazer"
              onClick={() => editor?.chain().focus().undo().run()}
            >
              <Undo2 size={20} />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Desfazer</TooltipContent>
        </Tooltip>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              disabled={!editor?.can().redo()}
              aria-label="Refazer"
              onClick={() => editor?.chain().focus().redo().run()}
            >
              <Redo2 size={20} />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Refazer</TooltipContent>
        </Tooltip>
      </div>

      <div ref={hostRef} className="editor-content" />
    </div>
  );
}

function TemplateDropdown({
  label,
  icon,
  templates,
  tooltip,
  onSelect,
}: {
  label: string;
  icon: React.ReactNode;
  templates: ControlTemplate[];
  tooltip: string;
  onSelect: (t: ControlTemplate) => void;
}) {
  return (
    <Tooltip>
      <DropdownMenu>
        <TooltipTrigger asChild>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm">
              {icon}
              {label}
              <ChevronDown size={18} />
            </Button>
          </DropdownMenuTrigger>
        </TooltipTrigger>
        <DropdownMenuContent>
          {templates.map((t) => (
            <DropdownMenuItem key={t.value} onSelect={() => onSelect(t)}>
              {t.label}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
      <TooltipContent>{tooltip}</TooltipContent>
    </Tooltip>
  );
}
