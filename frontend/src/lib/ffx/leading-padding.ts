/**
 * Separa o "padding de tela" do início de um texto canônico: a sequência
 * contínua de espaços, tabs, \n/\r e {TEXT_NEWLINE} (qualquer quantidade,
 * não apenas um token) com que o jogo alinha o texto na tela.
 *
 * Somente a UI usa isto (badge no editor / indicador no preview). O valor
 * canônico serializado continua incluindo o padding — nada é removido de
 * fato do dado.
 */

export interface LeadingPadding {
  /** Prefixo bruto (espaços, \n/\r, {TEXT_NEWLINE}) preservado. */
  padding: string;
  /** Texto sem o prefixo, para edição/exibição. */
  body: string;
}

const LEADING_PADDING_RE = /^(?:[\s]|\{TEXT_NEWLINE\})+/i;

export function splitLeadingPadding(raw: string): LeadingPadding {
  if (!raw) return { padding: '', body: '' };
  const match = raw.match(LEADING_PADDING_RE);
  if (!match) return { padding: '', body: raw };
  return { padding: match[0], body: raw.slice(match[0].length) };
}

export function summarizePadding(padding: string): string {
  if (!padding) return '';
  let newlines = 0;
  let spaces = 0;
  let i = 0;
  while (i < padding.length) {
    const rest = padding.slice(i).toUpperCase();
    if (rest.startsWith('{TEXT_NEWLINE}')) {
      newlines++;
      i += '{TEXT_NEWLINE}'.length;
      continue;
    }
    const ch = padding[i];
    if (ch === '\n' || ch === '\r') newlines++;
    else spaces++;
    i++;
  }
  const parts: string[] = [];
  if (newlines > 0) parts.push(`{TEXT_NEWLINE} ×${newlines}`);
  if (spaces > 0) parts.push(`␣ ×${spaces}`);
  return parts.join(' · ');
}
