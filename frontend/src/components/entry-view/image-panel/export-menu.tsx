'use client';

import { ChevronDown, Download, Save } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

/**
 * "Exportar como" do painel: escolhe o formato antes de abrir o seletor de
 * arquivo. O .dds é o padrão (sem perda, volta no repack); o .png é só para
 * visualizar/distribuir.
 */
export function ImageExportMenu({
  disabled,
  onSave,
}: {
  disabled: boolean;
  onSave: (format: 'dds' | 'png') => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          size="sm"
          variant="outline"
          disabled={disabled}
          title="Salvar como — o .dds é o padrão (sem perda, volta no repack); o .png é só para visualizar/distribuir"
        >
          <Save size={16} />
          Exportar
          <ChevronDown size={12} />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start">
        <DropdownMenuItem onClick={() => onSave('dds')}>
          <Save size={14} />
          .dds
          <span className="ml-2 text-xs opacity-60">padrão</span>
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => onSave('png')}>
          <Download size={14} />
          .png
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
