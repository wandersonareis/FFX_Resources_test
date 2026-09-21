import { GameTextParserService } from './game-text-parser.service';

describe('GameTextParserService', () => {
  let service: GameTextParserService;

  beforeEach(() => {
    service = new GameTextParserService();
  });

  it('converts colors to spans and back', () => {
    const raw = '{CLR:BLUE}Magia{CLR:WHITE} normal';
    const html = service.parseGameTextToHTML(raw);
    expect(html).toContain('color: #2563eb');
    expect(html).toContain('Magia');
  });

  it('converts italic tags to em', () => {
    const html = service.parseGameTextToHTML('{TEXT_ITALIC}it{TEXT_NORMAL}');
    expect(html).toContain('<em>it</em>');
  });

  it('splits real newlines and TEXT_NEWLINE into paragraphs', () => {
    const html = service.parseGameTextToHTML('a\nb{TEXT_NEWLINE}c');
    expect(html).toBe('<p>a</p><p>b</p><p>c</p>');
  });

  it('renders BUTTON/ICON as chips with last segment as label', () => {
    const html = service.parseGameTextToHTML('Pressione {BUTTON:32:CIRCLE} e {ICON:80:Red Gate}');
    expect(html).toContain('data-game-tag="{BUTTON:32:CIRCLE}"');
    expect(html).toContain('>CIRCLE</span>');
    expect(html).toContain('data-game-tag="{ICON:80:Red Gate}"');
    expect(html).toContain('>Red Gate</span>');
  });

  it('uses quoted last value as chip label', () => {
    const html = service.parseGameTextToHTML('{MCR:s06:l3F:"Yevon"}');
    expect(html).toContain('data-game-tag="{MCR:s06:l3F:&quot;Yevon&quot;}"');
    expect(html).toContain('>Yevon</span>');
  });

  it('labels PC chips with the character name', () => {
    const html = service.parseGameTextToHTML('{PC:01:YUNA}');
    expect(html).toContain('>YUNA</span>');
  });

  it('labels VAR chips as Variável and locks them', () => {
    const html = service.parseGameTextToHTML('{VAR:01}');
    expect(html).toContain('data-locked="true"');
    expect(html).toContain('>Variável</span>');
  });

  it('shows CMD/HEX/UNK chips as-is and locked', () => {
    const html = service.parseGameTextToHTML('{CMD:0E:40} {HEX:0B} {UNKCHR:FF}');
    expect(html).toContain('>{CMD:0E:40}</span>');
    expect(html).toContain('>{HEX:0B}</span>');
    expect(html).toContain('>{UNKCHR:FF}</span>');
    expect(html.match(/data-locked="true"/g)?.length).toBe(3);
  });

  it('round-trips formatting through Tiptap JSON', () => {
    const json = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            { type: 'text', marks: [{ type: 'textStyle', attrs: { color: '#dc2626' } }], text: 'Fogo' },
            { type: 'text', text: ' e ' },
            { type: 'text', marks: [{ type: 'italic' }], text: 'it' },
            { type: 'gameTag', attrs: { value: '{PC:01:YUNA}', label: 'YUNA', locked: false } },
            { type: 'gameTag', attrs: { value: '{VAR:01}', label: 'Variável', locked: true } },
          ],
        },
      ],
    };
    // Nota: TEXT_NORMAL fecha no fim do parágrafo (chips não participam
    // do sync de itálico); byte-idêntico após ParseCommand.
    expect(service.serializeJSONToGameText(json)).toBe(
      '{CLR:RED}Fogo{CLR:WHITE} e {TEXT_ITALIC}it{PC:01:YUNA}{VAR:01}{TEXT_NORMAL}'
    );
  });

  it('closes open color and italic at paragraph end', () => {
    const json = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            { type: 'text', marks: [{ type: 'textStyle', attrs: { color: '#2563eb' } }], text: 'Azul' },
          ],
        },
      ],
    };
    expect(service.serializeJSONToGameText(json)).toBe('{CLR:BLUE}Azul{CLR:WHITE}');
  });
});
