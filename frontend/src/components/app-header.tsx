'use client';

import { ChevronDown, Settings } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { ThemeToggle } from '@/components/theme-toggle';
import { SHORTCUTS, shortcutTip } from '@/lib/ffx/shortcuts';
import {
  EXPORT_FORMAT_LABELS,
  exportSelection,
  ExportFormat,
  useExportSelection,
} from '@/lib/ffx/export-selection';

/**
 * Barra superior (título + tema + configurações + formato de exportação).
 * Só precisa do callback de abrir Configurações: o formato vem da store
 * módulo `exportSelection`, então não atravessa props.
 */
export function AppHeader({
  onOpenConfig,
}: {
  onOpenConfig: () => void;
}) {
  const exportFormat = useExportSelection().format;

  return (
    <header className="sticky top-0 z-10 flex items-center gap-2 bg-chrome px-4 py-2 border-b">
      <span className="text-xl font-malva-medium text-gradient">
        FINAL FANTASY X and X-2 HD Remaster Resources Editor
      </span>
      <span className="flex-1" />
      <ThemeToggle />
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            onClick={onOpenConfig}
            aria-label="Configurações"
          >
            <Settings size={22} />
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          {shortcutTip(SHORTCUTS.config.label, SHORTCUTS.config.keys)}
        </TooltipContent>
      </Tooltip>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="sm" className="min-w-19 h-9 gap-1 px-2 text-sm">
            {EXPORT_FORMAT_LABELS[exportFormat]}
            <ChevronDown size={14} />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuLabel className="text-muted-foreground">
            Formato
          </DropdownMenuLabel>
          <DropdownMenuRadioGroup
            value={exportFormat}
            onValueChange={(value) => exportSelection.setFormat(value as ExportFormat)}
          >
            {(Object.keys(EXPORT_FORMAT_LABELS) as ExportFormat[]).map((format) => (
              <DropdownMenuRadioItem key={format} value={format}>
                {EXPORT_FORMAT_LABELS[format]}
              </DropdownMenuRadioItem>
            ))}
          </DropdownMenuRadioGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  );
}
