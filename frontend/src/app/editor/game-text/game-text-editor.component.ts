import {
  AfterViewInit,
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  ElementRef,
  NgZone,
  OnDestroy,
  effect,
  inject,
  input,
  output,
  viewChild,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatButtonModule } from '@angular/material/button';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatTooltipModule } from '@angular/material/tooltip';
import { Editor } from '@tiptap/core';
import StarterKit from '@tiptap/starter-kit';
import { TextStyle } from '@tiptap/extension-text-style';
import { Color } from '@tiptap/extension-color';
import { GameTagExtension } from './game-tag.extension';
import { createPasteHandlerExtension } from './paste-handler.extension';
import { LockedTagsExtension } from './locked-tags.extension';
import { GameTextParserService } from './game-text-parser.service';
import { KNOWN_COLORS, extractChipLabel, isLockedTagInner } from './game-text-tags';

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

@Component({
  selector: 'app-game-text-editor',
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule,
    MatButtonToggleModule,
    MatIconModule,
    MatMenuModule,
    MatTooltipModule,
  ],
  templateUrl: './game-text-editor.component.html',
  styleUrl: './game-text-editor.component.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class GameTextEditorComponent implements AfterViewInit, OnDestroy {
  private readonly parser = inject(GameTextParserService);
  private readonly zone = inject(NgZone);
  private readonly cdr = inject(ChangeDetectorRef);

  /** Texto canônico com tags (ex: linha de TextRow.text[lang]). */
  value = input<string>('');
  /** Emite o texto canônico serializado a cada edição. */
  valueChange = output<string>();

  editorHost = viewChild<ElementRef>('editorHost');

  protected editor: Editor | null = null;
  protected readonly colors = Object.values(KNOWN_COLORS).filter((c) => c.name !== 'WHITE');
  protected readonly buttonTemplates = BUTTON_TEMPLATES;
  protected readonly dpadTemplates = DPAD_TEMPLATES;
  protected readonly iconTemplates = ICON_TEMPLATES;
  protected readonly commandTemplates = COMMAND_TEMPLATES;

  private suppressUpdate = false;

  constructor() {
    effect(() => {
      const text = this.value();
      if (this.editor && !this.suppressUpdate) {
        const html = this.parser.parseGameTextToHTML(text);
        if (this.editor.getHTML() !== html) {
          this.editor.commands.setContent(html, { emitUpdate: false });
        }
      }
    });
  }

  ngAfterViewInit(): void {
    this.zone.runOutsideAngular(() => {
      this.editor = new Editor({
        element: this.editorHost()?.nativeElement,
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
          createPasteHandlerExtension(this.parser),
        ],
        content: this.parser.parseGameTextToHTML(this.value()),
        onUpdate: ({ editor }) => {
          this.zone.run(() => {
            this.suppressUpdate = true;
            try {
              this.valueChange.emit(this.parser.serializeJSONToGameText(editor.getJSON()));
            } finally {
              queueMicrotask(() => (this.suppressUpdate = false));
            }
            this.cdr.markForCheck();
          });
        },
        onSelectionUpdate: () => {
          this.zone.run(() => this.cdr.markForCheck());
        },
      });
    });
    this.cdr.markForCheck();
  }

  ngOnDestroy(): void {
    this.editor?.destroy();
    this.editor = null;
  }

  protected toggleItalic(): void {
    this.editor?.chain().focus().toggleItalic().run();
  }

  protected unsetItalic(): void {
    this.editor?.chain().focus().unsetItalic().run();
  }

  protected isItalicActive(): boolean {
    return this.editor?.isActive('italic') ?? false;
  }

  protected setColor(hex: string): void {
    this.editor?.chain().focus().setColor(hex).run();
  }

  protected resetColor(): void {
    this.editor?.chain().focus().unsetColor().run();
  }

  protected isColorActive(hex: string): boolean {
    return this.editor?.isActive('textStyle', { color: hex }) ?? false;
  }

  protected insertControlTag(template: ControlTemplate): void {
    if (!this.editor) return;
    const inner = template.value.slice(1, -1);
    this.editor
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
  }

  protected undo(): void {
    this.editor?.chain().focus().undo().run();
  }

  protected redo(): void {
    this.editor?.chain().focus().redo().run();
  }

  protected canUndo(): boolean {
    return this.editor?.can().undo() ?? false;
  }

  protected canRedo(): boolean {
    return this.editor?.can().redo() ?? false;
  }
}
