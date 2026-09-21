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
