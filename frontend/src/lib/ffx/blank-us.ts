// Texto 'us' em branco — espelho do isBlankText do backend
// (backend/services/import.service.go): vazio, só espaço ou o "-" que é o
// próprio texto do jogo. Export e import partem deste predicado; a tabela
// da UI mostra só o que o artefato exportado mostraria.

/**
 * Texto não traduzível no idioma padrão: vazio, só espaço, o hífen
 * "-" (placeholder do jogo, ~metade do data/ pristine), ou só tags de
 * formatação pura ({TEXT_NEWLINE}, {TEXT_ITALIC}, CLR:/COLOR:...) — que é o
 * tipo de row o editor nem deveria mostrar. Espelho do isBlankText do backend
 * (backend/services/import.service.go, via converter.IsTextTag).
 */
export function isBlankText(t: string | undefined | null): boolean {
  const s = stripTextTags(t ?? '').trim();
  return s === '' || s === '-';
}

// stripTextTags remove do texto todas as tags classificadas como text-tag
// (TEXT_ITALIC/TEXT_NORMAL/\n/TEXT_NEWLINE/CLR:/COLOR:), que não carregam
// conteúdo — ausência delas não é divergência. Tags de controle (ícone,
// pausa…) ficam: elas têm significado. Igual ao backend.
function stripTextTags(t: string): string {
  const runes = [...t];
  const out: string[] = [];
  for (let i = 0; i < runes.length; i++) {
    if (runes[i] !== '{') {
      out.push(runes[i]);
      continue;
    }
    let end = -1;
    for (let j = i + 1; j < runes.length; j++) {
      if (runes[j] === '}') {
        end = j;
        break;
      }
    }
    if (end < 0) {
      out.push(runes[i]);
      continue;
    }
    const cmd = runes.slice(i + 1, end).join('');
    if (isTextTag(cmd)) {
      i = end;
      continue;
    }
    out.push(runes[i]);
  }
  return out.join('');
}

function isTextTag(cmd: string): boolean {
  return (
    cmd === 'TEXT_ITALIC' ||
    cmd === 'TEXT_NORMAL' ||
    cmd === '\\n' ||
    cmd === 'TEXT_NEWLINE' ||
    cmd.startsWith('CLR:') ||
    cmd.startsWith('COLOR:')
  );
}

/**
 * Row da view sem 'us' utilizável — fica fora da TABELA e da contagem de
 * progresso do badge, sempre no idioma padrão (o backend exporta só o que
 * tem 'us': se o 'us' não existe, os outros idiomas não interessam).
 */
export function isBlankSourceRow(text: string | undefined | null): boolean {
  return isBlankText(text);
}
