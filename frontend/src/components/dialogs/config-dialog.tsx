'use client';

import { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { FolderOpen, X } from 'lucide-react';
import {
  GetGameFilesLocation,
  GetTranslateLocation,
  SelectDirectory,
} from '@/wailsjs/go/main/App';
import { EventsEmit } from '@/wailsjs/runtime/runtime';
import { useWailsEvent } from '@/lib/ffx/use-wails-event';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

type LocationKey = 'GameFilesLocation' | 'TranslateLocation';

interface DirectoryInput {
  key: LocationKey;
  eventName: string;
  label: string;
  hint: string;
  dialogTitle: string;
  value: string;
}

/**
 * Diálogo de diretórios: apenas gamefiles e translated (fonte do
 * reimport, <game>/mods/translated por padrão). Extract/reimport são
 * internos do backend e não aparecem aqui.
 * Toda mudança é aplicada imediatamente (ao confirmar o campo ou
 * escolher no seletor): o backend persiste no config.json e re-emite
 * o valor + Refresh_Tree.
 */
export function ConfigDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const [gameDirectory, setGameDirectory] = useState('');
  const [translatedDirectory, setTranslatedDirectory] = useState('');

  useWailsEvent('GameFilesLocation', (data) => setGameDirectory((data as string) ?? ''));
  useWailsEvent('TranslateLocation', (data) => setTranslatedDirectory((data as string) ?? ''));

  useEffect(() => {
    if (!open) return;
    GetGameFilesLocation()
      .then((v) => setGameDirectory(v ?? ''))
      .catch(notify);
    GetTranslateLocation()
      .then((v) => setTranslatedDirectory(v ?? ''))
      .catch(notify);
  }, [open]);

  const inputs: DirectoryInput[] = [
    {
      key: 'GameFilesLocation',
      eventName: 'GameLocationChanged',
      label: 'Arquivos originais do jogo',
      hint: 'Padrão: <execução>/data',
      dialogTitle: 'Selecione a pasta dos arquivos originais do jogo',
      value: gameDirectory,
    },
    {
      key: 'TranslateLocation',
      eventName: 'TranslateLocationChanged',
      label: 'Arquivos traduzidos (reimport)',
      hint: 'Padrão: <jogo>/mods/translated',
      dialogTitle: 'Selecione a pasta dos arquivos traduzidos',
      value: translatedDirectory,
    },
  ];

  const apply = (item: DirectoryInput, raw: string) => {
    const path = (raw ?? '').trim();
    if (!path) {
      toast('Informe um diretório válido.', { duration: 3000 });
      return;
    }
    if (path === item.value) return;
    try {
      EventsEmit(item.eventName, path);
    } catch (error) {
      notify(error);
    }
  };

  const selectDirectory = async (item: DirectoryInput) => {
    try {
      const path = await SelectDirectory(item.dialogTitle);
      if (path) apply(item, path);
    } catch (error) {
      notify(error);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[640px]">
        <DialogHeader>
          <DialogTitle>Configurações</DialogTitle>
          <DialogDescription className="sr-only">Diretórios usados pelo app</DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-3 pt-2 min-w-0">
          {inputs.map((item) => (
            <div key={item.key} className="flex items-center gap-2">
              <div className="grid flex-1 gap-1.5">
                <Label htmlFor={item.key}>{item.label}</Label>
                <Input
                  id={item.key}
                  value={item.value}
                  onChange={(e) => apply(item, e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') apply(item, (e.target as HTMLInputElement).value);
                  }}
                />
                <p className="text-xs text-muted-foreground">{item.hint}</p>
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => void selectDirectory(item)}
                aria-label={`Selecionar ${item.label}`}
                title="Procurar pasta"
              >
                <FolderOpen size={20} />
              </Button>
            </div>
          ))}
        </div>

        <DialogFooter>
          <Button onClick={() => onOpenChange(false)}>
            <X size={18} />
            Fechar
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function notify(error: unknown): void {
  const message = error instanceof Error ? error.message : String(error ?? 'Erro');
  toast.error(message, { duration: 4000 });
  EventsEmit('Notify', { severity: 'error', message });
}
