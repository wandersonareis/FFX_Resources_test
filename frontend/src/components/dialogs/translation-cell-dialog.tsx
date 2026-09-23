'use client';

import { useState } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { dto } from '@/wailsjs/go/models';
import { SOURCE_LANG } from '@/lib/ffx/save-all';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { GameTextEditor } from '@/components/editor/game-text-editor';
import { GameTextView } from '@/components/game-text-view';

export interface TranslationCellDialogProps {
  open: boolean;
  row: dto.TextRow | null;
  /** Idiomas de consulta, sem o idioma de origem (us). */
  languages: Array<{ code: string; name: string }>;
  /** Há linha anterior no ARQUIVO atual (navegação nunca troca de arquivo). */
  hasPrevious: boolean;
  /** Há próxima linha no ARQUIVO atual. */
  hasNext: boolean;
  /**
   * Navegação sem fechar o modal. value só vem preenchido quando há edição
   * não salva (aplicada como rascunho antes de trocar de linha).
   */
  onNavigate: (direction: 'prev' | 'next', value?: string) => void;
  /** Fechamento: value != undefined aplica a edição. */
  onClosed: (value?: string) => void;
}

export function TranslationCellDialog({
  open,
  row,
  languages,
  hasPrevious,
  hasNext,
  onNavigate,
  onClosed,
}: TranslationCellDialogProps) {
  // Remontado por key no pai a cada linha: o state inicial já reflete a linha.
  const [text, setText] = useState(row?.text?.[SOURCE_LANG] ?? '');

  const original = row?.text?.[SOURCE_LANG] ?? '';

  const changed = text !== original;
  const referenceLangs = languages.filter((l) => l.code !== SOURCE_LANG);

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!o) onClosed();
      }}
    >
      <DialogContent className="sm:max-w-[720px] max-h-[90vh] flex flex-col pt-10">
        {/* Linha do título + navegação, abaixo do botão X de fechar */}
        <div className="flex items-center justify-between gap-2">
          <DialogTitle className="truncate">
            Linha #{row?.index}
            {row?.name ? <span className="opacity-70"> · {row.name}</span> : null}
          </DialogTitle>
          <div className="flex shrink-0 items-center gap-1.5">
            {hasPrevious ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onNavigate('prev', changed ? text : undefined)}
              >
                <ChevronLeft size={16} />
                Anterior
              </Button>
            ) : null}
            {hasNext ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onNavigate('next', changed ? text : undefined)}
              >
                Próxima
                <ChevronRight size={16} />
              </Button>
            ) : null}
          </div>
        </div>
        <DialogDescription className="mt-1">
          Tradução ({SOURCE_LANG}) — gravada no jogo ao salvar
        </DialogDescription>

        <div className="flex flex-col gap-4 overflow-hidden">
          <div className="grid gap-2.5 overflow-y-auto flex-1 min-h-[120px] max-h-[42vh] pr-1">
            {referenceLangs.length === 0 ? (
              <p className="opacity-70">Sem outros idiomas nesta linha.</p>
            ) : (
              referenceLangs.map((lang) => (
                <div key={lang.code}>
                  <span className="text-xs font-semibold opacity-75">
                    {lang.name} ({lang.code})
                  </span>
                  <GameTextView
                    className="mt-0.5"
                    text={row?.text?.[lang.code]}
                    fallback="—"
                  />
                </div>
              ))
            )}
          </div>

          <div className="grid gap-1.5">
            <span className="text-xs font-semibold opacity-75">
              Tradução ({SOURCE_LANG}) — gravada no jogo ao salvar
            </span>
            <GameTextEditor value={text} onValueChange={setText} />
          </div>
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => onClosed()}>
            Cancelar
          </Button>
          <Button disabled={!changed} onClick={() => onClosed(text)}>
            Salvar
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
