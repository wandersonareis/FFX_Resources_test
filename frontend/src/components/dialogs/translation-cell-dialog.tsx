'use client';

import { useState } from 'react';
import { dto } from '@/wailsjs/go/models';
import { SOURCE_LANG } from '@/lib/ffx/save-all';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { GameTextEditor } from '@/components/editor/game-text-editor';

export interface TranslationCellDialogProps {
  open: boolean;
  row: dto.TextRow | null;
  /** Idiomas de consulta, sem o idioma de origem (us). */
  languages: Array<{ code: string; name: string }>;
  /** Fechamento: value != undefined aplica a edição. */
  onClosed: (value?: string) => void;
}

export function TranslationCellDialog({
  open,
  row,
  languages,
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
      <DialogContent className="sm:max-w-[720px] max-h-[90vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>
            Linha #{row?.index}
            {row?.name ? <span className="opacity-70"> · {row.name}</span> : null}
          </DialogTitle>
          <DialogDescription>
            Tradução ({SOURCE_LANG}) — gravada no jogo ao salvar
          </DialogDescription>
        </DialogHeader>

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
                  <p className="mt-0.5 whitespace-pre-wrap">
                    {row?.text?.[lang.code] ?? '—'}
                  </p>
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
