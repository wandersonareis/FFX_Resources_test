"use client";

import { useEffect, useState } from "react";
import { loggedToast as toast } from "@/lib/ffx/toast-logged";
import { FolderOpen, X } from "lucide-react";
import {
  GetEnableMods,
  GetGameExeLocation,
  GetGameFilesLocation,
  GetTranslateLocation,
  SelectDirectory,
  SelectGameExeFile,
  SetEnableMods,
} from "@/wailsjs/go/main/App";
import { EventsEmit } from "@/wailsjs/runtime/runtime";
import { useWailsEvent } from "@/lib/ffx/use-wails-event";
import { invalidateVbfCache } from "@/lib/ffx/vbf";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type LocationKey = "GameFilesLocation" | "TranslateLocation";

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
  const [gameDirectory, setGameDirectory] = useState("");
  const [translatedDirectory, setTranslatedDirectory] = useState("");
  const [gameExe, setGameExe] = useState("");
  const [enableMods, setEnableMods] = useState(true);

  useWailsEvent("GameFilesLocation", (data) =>
    setGameDirectory((data as string) ?? ""),
  );
  useWailsEvent("TranslateLocation", (data) =>
    setTranslatedDirectory((data as string) ?? ""),
  );
  useWailsEvent("GameExeLocation", (data) =>
    setGameExe((data as string) ?? ""),
  );

  useEffect(() => {
    if (!open) return;
    GetGameFilesLocation()
      .then((v) => setGameDirectory(v ?? ""))
      .catch(notify);
    GetTranslateLocation()
      .then((v) => setTranslatedDirectory(v ?? ""))
      .catch(notify);
    GetGameExeLocation()
      .then((v) => setGameExe(v ?? ""))
      .catch(notify);
    GetEnableMods()
      .then((v) => setEnableMods(Boolean(v)))
      .catch(notify);
  }, [open]);

  const applyEnableMods = async (checked: boolean) => {
    const previous = enableMods;
    setEnableMods(checked);
    try {
      await SetEnableMods(checked);
    } catch (error) {
      setEnableMods(previous);
      notify(error);
    }
  };

  const inputs: DirectoryInput[] = [
    {
      key: "GameFilesLocation",
      eventName: "GameLocationChanged",
      label: "Arquivos originais do jogo",
      hint: "Padrão: <execução>/data",
      dialogTitle: "Selecione a pasta dos arquivos originais do jogo",
      value: gameDirectory,
    },
    {
      key: "TranslateLocation",
      eventName: "TranslateLocationChanged",
      label: "Arquivos traduzidos (reimport)",
      hint: "Padrão: <jogo>/mods/translated",
      dialogTitle: "Selecione a pasta dos arquivos traduzidos",
      value: translatedDirectory,
    },
  ];

  const apply = (item: DirectoryInput, raw: string) => {
    const path = (raw ?? "").trim();
    if (!path) {
      toast("Informe um diretório válido.", { duration: 3000 });
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

  /**
   * O executável é um ARQUIVO (fica fora do map de Locations) e é dele que
   * saem as raízes dos .vbf (pasta do exe e pasta data/). Commite só no
   * Enter/blur/seletor: gravar a cada tecla fecharia os containers abertos
   * a cada passo. O backend re-emite o valor e recarrega a árvore.
   */
  const commitExe = (raw: string) => {
    const path = (raw ?? "").trim();
    if (path === gameExe) return;
    setGameExe(path);
    // O executável mudou: o cache local de diretórios/entradas do .vbf
    // aponta para containers que podem nem existir mais.
    invalidateVbfCache();
    try {
      EventsEmit("GameExeLocationChanged", path);
    } catch (error) {
      notify(error);
    }
  };

  const selectExe = async () => {
    try {
      const path = await SelectGameExeFile();
      if (path) commitExe(path);
    } catch (error) {
      notify(error);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-160">
        <DialogHeader>
          <DialogTitle>Configurações</DialogTitle>
          <DialogDescription className="sr-only">
            Diretórios usados pelo app
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-3 pt-2 min-w-0">
          <div className="flex items-center gap-2 pt-2">
            <div className="grid flex-1 gap-1.5">
              <Label htmlFor="GameExeLocation">Executável do jogo</Label>
              <Input
                id="GameExeLocation"
                value={gameExe}
                placeholder="FFX.exe ou FFX-2.exe"
                onChange={(e) => setGameExe(e.target.value)}
                onBlur={() => commitExe(gameExe)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") commitExe(gameExe);
                }}
              />
              <p className="text-xs text-muted-foreground">
                Dele saem as raízes dos containers <code>.vbf</code> (pasta do
                exe e pasta <code>data/</code>) exibidos no fim da árvore.
              </p>
            </div>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => void selectExe()}
              aria-label="Selecionar o executável do jogo"
              title="Procurar arquivo"
            >
              <FolderOpen size={20} />
            </Button>
          </div>
          {inputs.map((item) => (
            <div key={item.key} className="flex items-center gap-2">
              <div className="grid flex-1 gap-1.5">
                <Label htmlFor={item.key}>{item.label}</Label>
                <Input
                  id={item.key}
                  value={item.value}
                  onChange={(e) => apply(item, e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter")
                      apply(item, (e.target as HTMLInputElement).value);
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

        <div className="flex items-start gap-2 pt-2">
          <Checkbox
            id="EnableMods"
            checked={enableMods}
            onCheckedChange={(v) => void applyEnableMods(Boolean(v))}
          />
          <div className="grid gap-1 leading-none">
            <Label htmlFor="EnableMods" className="cursor-pointer">
              Preferir binários traduzidos (mods)
            </Label>
            <p className="text-xs text-muted-foreground">
              Ao carregar, usa primeiro os binários salvos em
              <code className="mx-1">mods/</code>e cai para os originais quando
              não existirem — a tradução continua de onde parou. O import de
              JSON/strings grava os binários nessa pasta.
            </p>
          </div>
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
  const message =
    error instanceof Error ? error.message : String(error ?? "Erro");
  toast.error(message, { duration: 4000 });
  EventsEmit("Notify", { severity: "error", message });
}
