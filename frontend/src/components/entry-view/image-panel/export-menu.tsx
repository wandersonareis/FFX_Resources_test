'use client';

import { ChevronDown, Download, FolderInput, Save } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

/**
 * "Exportar como" do painel: dois destinos diferentes no mesmo menu.
 *
 *  - `.dds` / `.png`  → "salvar como" (selectImageSavePath): destino LIVRE
 *    no disco, UM formato, imagem aberta.
 *  - `.dds + .png`    → extração para a pasta de trabalho `mods/images/`
 *    (mesmo fluxo do antigo botão "Extrair .dds + .png"): grava OS DOIS
 *    formatos com o caminho interno preservado; em data/ abre o diálogo de
 *    alcance (só esta / todas as cópias), em .vbf extrai direto.
 *
 * O .dds é o padrão sem perda (volta no repack); o .png é só para
 * visualizar/distribuir.
 */
export function ImageExportMenu({
  disabled,
  readOnly,
  onSave,
  onExtract,
}: {
  disabled: boolean;
  /** true = textura veio do .vbf (extração não abre diálogo de alcance). */
  readOnly: boolean;
  onSave: (format: 'dds' | 'png') => void;
  onExtract: () => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          size="sm"
          variant="outline"
          disabled={disabled}
          title="Exportar — salvar como .dds/.png em pasta livre, ou extrair os dois formatos para mods/"
        >
          <Save size={16} />
          Exportar
          <ChevronDown size={12} />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="start"
        className="min-w-44"
      >
        <DropdownMenuItem onClick={() => onSave('dds')}>
          <Save size={14} />
          .dds
          <span className="ml-auto pl-3 text-xs opacity-60">padrão</span>
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => onSave('png')}>
          <Download size={14} />
          .png
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={onExtract}
          title={
            readOnly
              ? 'Extrai esta imagem do .vbf como .dds e .png, preservando o caminho interno (mods/images)'
              : 'Gera .dds e .png na pasta de imagens de trabalho — o diálogo pergunta se é só esta ou todas as cópias'
          }
        >
          <FolderInput size={14} />
          .dds + .png
          <span className="ml-auto pl-3 text-xs opacity-60">mods/</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
