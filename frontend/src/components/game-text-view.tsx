'use client';

import { useMemo } from 'react';
import { formatGameTextForDisplay } from '@/lib/ffx/game-text-format';

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
 * Casca fina sobre `formatGameTextForDisplay` (formatador compartilhado com o
 * modal de busca). O parser escapa todo texto/atributo, então o HTML é seguro
 * para dangerouslySetInnerHTML.
 */
export function GameTextView({ text, fallback = '', className }: GameTextViewProps) {
  const { paddingSummary, html } = useMemo(
    () => formatGameTextForDisplay(text),
    [text]
  );

  if (!text) {
    return fallback ? <div className={className}>{fallback}</div> : null;
  }

  return (
    <div className={className}>
      {paddingSummary !== '' && (
        <div className="mb-1 flex items-center gap-1 text-xs text-muted-foreground">
          <span>␍ padding de tela:</span>
          <code className="rounded bg-muted px-1">
            {paddingSummary}
          </code>
        </div>
      )}
      <div
        className="[&_p]:mb-1 [&_p:last-child]:mb-0"
        dangerouslySetInnerHTML={{ __html: html }}
      />
    </div>
  );
}
