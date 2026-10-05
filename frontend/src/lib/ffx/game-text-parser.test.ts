import { describe, expect, it } from 'bun:test';
import { gameTextParser } from './game-text-parser';

// O editor compara parseGameTextToHTML(value) com editor.getHTML() (literal)
// e só chama setContent quando diferem — HTML que não bate com o renderHTML do
// nó gameTag ou com a serialização de texto do browser faria o cursor cair à
// toa. Estes testos fixam essa ordem/escapamento.
describe('parseGameTextToHTML: chips', () => {
  it('emite os atributos na mesma ordem do renderHTML', () => {
    const html = gameTextParser.parseGameTextToHTML('x{CMD:01:02}y');
    expect(html).toBe(
      '<p>x<span data-game-tag="{CMD:01:02}" class="game-tag-chip locked" ' +
        'data-locked="true">{CMD:01:02}</span>y</p>'
    );
  });

  it('marca o chip inválido com data-invalid no fim', () => {
    const html = gameTextParser.parseGameTextToHTML(
      'Oi {PC:ZZ:YUNA}',
      () => false
    );
    expect(html).toBe(
      '<p>Oi <span data-game-tag="{PC:ZZ:YUNA}" class="game-tag-chip invalid" ' +
        'data-invalid="true">YUNA</span></p>'
    );
  });

  it('chip válido não ganha data-locked/data-invalid', () => {
    const html = gameTextParser.parseGameTextToHTML(
      'Oi {PC:01:YUNA}',
      () => true
    );
    expect(html).toBe(
      '<p>Oi <span data-game-tag="{PC:01:YUNA}" class="game-tag-chip">YUNA</span></p>'
    );
  });

  it('sem validador nenhum chip é invalid', () => {
    const html = gameTextParser.parseGameTextToHTML('{HEX:DEADBEEF}');
    expect(html).toContain('class="game-tag-chip locked"');
    expect(html).not.toContain('data-invalid');
  });

  it('chip BUTTON com glifo vira o ícone direto no texto (sem chip)', () => {
    const html = gameTextParser.parseGameTextToHTML('{BUTTON:31:X}');
    expect(html).toBe(
      '<p><span data-game-tag="{BUTTON:31:X}">' +
        '<span class="gb-sprite gb-a" title="X"></span></span></p>'
    );
  });

  it('chips de direção viram SEQUÊNCIA de setas aninhadas ao wrapper (sem chip)', () => {
    const html = gameTextParser.parseGameTextToHTML('{BUTTON:43:Direcional Up+Right}');
    expect(html).toBe(
      '<p>' +
        '<span data-game-tag="{BUTTON:43:Direcional Up+Right}">' +
        '<span class="gb-sprite gb-arrow-up" title="Direcional Up+Right"></span>' +
        '<span class="gb-sprite gb-arrow-right" title="Direcional Up+Right"></span>' +
        '</span></p>'
    );
    // Triplas: 3 setas aninhadas na mesma ordem
    const html3 = gameTextParser.parseGameTextToHTML('{BUTTON:4B:Direcional Up+Left+Right}');
    expect((html3.match(/gb-arrow/g) ?? []).length).toBe(3);
    expect(html3.indexOf('gb-arrow-up')).toBeLessThan(html3.indexOf('gb-arrow-left'));
    expect(html3.indexOf('gb-arrow-left')).toBeLessThan(html3.indexOf('gb-arrow-right'));
    expect(html3).not.toContain('game-tag-chip');
  });

  it('chips All e Direcional genérico usam o cursor ✛', () => {
    for (const tag of ['{BUTTON:4F:Direcional All}', '{BUTTON:40:Direcional}']) {
      const html = gameTextParser.parseGameTextToHTML(tag);
      expect(html).toContain(
        `<span data-game-tag="${tag}"><span class="gb-sprite gb-cursor"`
      );
      expect(html).not.toContain('game-tag-chip');
    }
  });

  it('chip BUTTON sem glifo (Dummy) segue textual', () => {
    const html = gameTextParser.parseGameTextToHTML('{BUTTON:2D:Dummy}');
    expect(html).toContain('>Dummy</span>');
    expect(html).not.toContain('gb-sprite');
  });
});

describe('parseGameTextToHTML: escapamento', () => {
  it('aspas no texto ficam literais (igual à serialização do browser)', () => {
    expect(gameTextParser.parseGameTextToHTML('Ele disse "oi" & saiu')).toBe(
      '<p>Ele disse "oi" &amp; saiu</p>'
    );
  });

  it('apóstrofo no rótulo do chip fica literal', () => {
    const html = gameTextParser.parseGameTextToHTML("{MCR:s01:l05:\"Yevon's\"}");
    expect(html).toContain(">Yevon's</span>");
    expect(html).not.toContain('&#039;');
  });

  it('aspas dentro da tag escapam no atributo', () => {
    const html = gameTextParser.parseGameTextToHTML('{MCR:s01:l05:"Hi"}');
    expect(html).toContain('data-game-tag="{MCR:s01:l05:&quot;Hi&quot;}"');
  });

  it('escapa < e > do texto', () => {
    expect(gameTextParser.parseGameTextToHTML('a < b > c')).toBe(
      '<p>a &lt; b &gt; c</p>'
    );
  });
});

describe('parseGameTextToHTML: formatação não vira chip', () => {
  it('CLR/COLOR/TEXT_* ficam como marca, sem data-game-tag', () => {
    const html = gameTextParser.parseGameTextToHTML(
      '{CLR:RED}oi{TEXT_ITALIC}!{TEXT_NORMAL}{TEXT_NEWLINE}fim'
    );
    expect(html).not.toContain('data-game-tag');
    expect(html.split('<p>').length - 1).toBe(2);
  });
});

describe('serializeJSONToGameText', () => {
  it('emite o valor do chip verbatim', () => {
    const json = {
      type: 'doc',
      content: [
        {
          type: 'paragraph',
          content: [
            { type: 'text', text: 'Hi ' },
            { type: 'gameTag', attrs: { value: '{CMD:01:02}' } },
          ],
        },
      ],
    };
    expect(gameTextParser.serializeJSONToGameText(json)).toBe('Hi {CMD:01:02}');
  });
});
