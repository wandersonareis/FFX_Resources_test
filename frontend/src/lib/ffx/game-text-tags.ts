// Tags canônicas do backend (backend/core/converter/string_bytes.go +
// string_to_bytes.go). O diretório backend/core/encoding/tags é legado e
// NÃO é usado aqui. Terceiro campo das tags ({X:YY:Nome}) é só humano.

export interface ColorDefinition {
  label: string;
  name: string;
  hex: string;
}

// Nomes canônicos em MAIÚSCULAS, como emitidos por GetColorString.
// Hexes reaproveitados do protótipo tiptap-color-tag-editor.
export const KNOWN_COLORS: Record<string, ColorDefinition> = {
  BLUE: { label: 'Azul', name: 'BLUE', hex: '#2563eb' },
  RED: { label: 'Vermelho', name: 'RED', hex: '#dc2626' },
  YELLOW: { label: 'Amarelo', name: 'YELLOW', hex: '#eab308' },
  GREY: { label: 'Cinza', name: 'GREY', hex: '#6b7280' },
  PINK: { label: 'Rosa', name: 'PINK', hex: '#ec4899' },
  OL_PURPLE: { label: 'Roxo', name: 'OL_PURPLE', hex: '#9333ea' },
  OL_CYAN: { label: 'Ciano', name: 'OL_CYAN', hex: '#06b6d4' },
  WHITE: { label: 'Branco (reset)', name: 'WHITE', hex: '#ffffff' },
};

export const COLOR_RESET_TAG = 'WHITE';

/**
 * Normaliza um valor de cor (hex ou nome) para o nome canônico da tag.
 * Usado na serialização: mark textStyle.color -> {CLR:NOME}.
 */
export function getColorTagName(colorValue?: string | null): string | null {
  if (!colorValue) return null;
  const val = colorValue.trim().toLowerCase();
  for (const [key, def] of Object.entries(KNOWN_COLORS)) {
    if (
      key.toLowerCase() === val ||
      def.hex.toLowerCase() === val ||
      def.name.toLowerCase() === val
    ) {
      return def.name;
    }
  }
  if (val.startsWith('#')) return val.toUpperCase();
  return val.toUpperCase();
}

/**
 * Resolve o nome da tag de cor para CSS. WHITE = reset.
 * Retorna null se não for tag de cor.
 */
export function getResolvedColorFromTag(
  tagName: string
): { hex: string | null; isReset: boolean } | null {
  const normalized = tagName.trim().toUpperCase();
  if (normalized === 'WHITE') return { hex: null, isReset: true };
  if (KNOWN_COLORS[normalized]) {
    return { hex: KNOWN_COLORS[normalized].hex, isReset: false };
  }
  if (/^#[0-9A-F]{3}([0-9A-F]{3})?$/.test(normalized)) {
    return { hex: normalized, isReset: false };
  }
  return null;
}

// Prefixos cujos chips são BLOQUEADOS (valores calculados no jogo,
// nunca editáveis): CMD, HEX e qualquer UNK*.
const LOCKED_ASIS_PREFIXES = ['CMD:', 'HEX:', 'UNKCHR:', 'UNKDBLCHR:', 'UNKTPLCHR:'];

/**
 * Indica se a tag é bloqueada (CMD, HEX, UNK ou VAR).
 */
export function isLockedTagInner(inner: string): boolean {
  const upper = inner.trim().toUpperCase();
  if (upper === 'VAR' || upper.startsWith('VAR:')) return true;
  return LOCKED_ASIS_PREFIXES.some((p) => upper.startsWith(p));
}

/**
 * Extrai o label do chip a partir do conteúdo interno da tag:
 * - CMD/HEX/UNK*: exibe como está (label = tag completa).
 * - VAR: exibe apenas "Variável".
 * - Demais: último segmento após ':', sem aspas duplas.
 *   ex: MCR:s06:l3F:"Yevon" -> Yevon | PC:01:YUNA -> YUNA | PAUSE -> PAUSE
 */
export function extractChipLabel(inner: string): string {
  const trimmed = inner.trim();
  const upper = trimmed.toUpperCase();
  if (LOCKED_ASIS_PREFIXES.some((p) => upper.startsWith(p))) {
    return `{${trimmed}}`;
  }
  if (upper === 'VAR' || upper.startsWith('VAR:')) {
    return 'Variável';
  }
  const segments = trimmed.split(':');
  const last = segments[segments.length - 1].trim();
  const unquoted = last.replace(/^"|"$/g, '').trim();
  return unquoted || trimmed;
}

/**
 * Classes CSS dos tiles na sprite `public/pad_icon.png` (glifos do Xbox,
 * como o jogo renderiza no PC) para tags {BUTTON:XX:Nome}. Espelha o
 * buttonMap do backend (encoding_players.go):
 *   30 TRIANGLE→Y · 31 X→A · 32 CIRCLE→B · 33 SQUARE→X (layout Xbox);
 *   34–37→LB/RB/LT/RT · 38 START→▶ · 39 SELECT→◀.
 * Os direcionais (0x40–0x4F) viram SEQUÊNCIAS de setas na ordem do nome do
 * código (same do jogo): Up+Right = ↑,→. "Direcional" (0x40) e "All"
 * (0x4F) usam o cursor ✛ do pad (sem setas = todas as direções).
 *
 * Códigos sem glifo (Dummy/Dummy2, 0x001+) devolvem [] → chip textual.
 */
const BUTTON_CODE_SPRITES: Record<number, string[]> = {
  0x20: ['gb-lb'], // ?L1 (SWITCH)
  0x30: ['gb-y'],
  0x31: ['gb-a'],
  0x32: ['gb-b'],
  0x33: ['gb-x'],
  0x34: ['gb-lb'],
  0x35: ['gb-rb'],
  0x36: ['gb-lt'],
  0x37: ['gb-rt'],
  0x38: ['gb-start'],
  0x39: ['gb-back'],
  0x40: ['gb-cursor'], // Direcional (generico: cruz de direções)
  0x41: ['gb-arrow-up'],
  0x42: ['gb-arrow-right'],
  0x43: ['gb-arrow-up', 'gb-arrow-right'],
  0x44: ['gb-arrow-down'],
  0x45: ['gb-arrow-up', 'gb-arrow-down'],
  0x46: ['gb-arrow-down', 'gb-arrow-right'],
  0x47: ['gb-arrow-up', 'gb-arrow-right', 'gb-arrow-down'],
  0x48: ['gb-arrow-left'],
  0x49: ['gb-arrow-up', 'gb-arrow-left'],
  0x4a: ['gb-arrow-left', 'gb-arrow-right'],
  0x4b: ['gb-arrow-up', 'gb-arrow-left', 'gb-arrow-right'],
  0x4c: ['gb-arrow-left', 'gb-arrow-down'],
  0x4d: ['gb-arrow-up', 'gb-arrow-left', 'gb-arrow-down'],
  0x4e: ['gb-arrow-left', 'gb-arrow-down', 'gb-arrow-right'],
  0x4f: ['gb-cursor'], // Direcional All
};

/**
 * Sequência de classes de sprite para o chip de {BUTTON:XX:…}; lista vazia
 * quando o código não tem glifo (chip segue textual). Aceita tanto o inner
 * ("BUTTON:41:…", do parser) quanto a tag completa com chaves (attrs.value
 * do node no renderHTML).
 */
export function buttonSpriteClasses(inner: string): string[] {
  const match = /^BUTTON:([0-9A-Fa-f]{1,2})(?::|$)/i.exec(
    inner.trim().replace(/^\{|\}$/g, '')
  );
  if (!match) return [];
  const code = parseInt(match[1], 16);
  if (Number.isNaN(code)) return [];
  return BUTTON_CODE_SPRITES[code] ?? [];
}
