'use client';

import { useMemo } from 'react';
import { gameTextParser } from '@/lib/ffx/game-text-parser';

interface GameTextViewProps {
  /** Texto canônico com tags (ex: linha de TextRow.text[lang]). */
  text?: string;
  /** Conteúdo quando o texto está vazio (padrão: nada). */
  fallback?: string;
  className?: string;
}

/**
 * Exibição somente-leitura do texto do jogo, com a MESMA conversão do editor
 * ({TEXT_NEWLINE}/\n -> parágrafo, {CLR:…} -> cor, {TEXT_ITALIC} -> itálico,
 * demais tags -> chip). Evita mostrar tags cruas ao lado do texto normalizado.
 *
 * O parser escapa todo texto/atributo, então o HTML é seguro para
 * dangerouslySetInnerHTML.
 */
export function GameTextView({ text, fallback = '', className }: GameTextViewProps) {
  const html = useMemo(() => gameTextParser.parseGameTextToHTML(text ?? ''), [text]);

  if (!text) {
    return fallback ? <div className={className}>{fallback}</div> : null;
  }

  return (
    <div
      className={`game-text-view${className ? ` ${className}` : ''}`}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
