'use client';

import { Moon, Sun } from 'lucide-react';
import { useTheme } from 'next-themes';
import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip';

/**
 * Toggle light/dark do app. Alterna a classe `.dark` no <html> via
 * next-themes (persistida em localStorage). Renderiza um ícone só depois
 * da hidratação para não divergir entre servidor e cliente (o tema salvo
 * só existe no navegador).
 */
export function ThemeToggle() {
  const { resolvedTheme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => setMounted(true), []);

  const isDark = mounted && resolvedTheme === 'dark';

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          aria-label={isDark ? 'Ativar tema claro' : 'Ativar tema escuro'}
          onClick={() => setTheme(isDark ? 'light' : 'dark')}
        >
          {isDark ? <Sun size={20} /> : <Moon size={20} />}
        </Button>
      </TooltipTrigger>
      <TooltipContent>
        {isDark ? 'Tema claro' : 'Tema escuro'}
      </TooltipContent>
    </Tooltip>
  );
}
