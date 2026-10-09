import { createColumnHelper, tableFeatures } from '@tanstack/react-table';
import { dto } from '@/wailsjs/go/models';

// Infra compartilhada da tabela (TanStack v9: features + columnHelper).
export const features = tableFeatures({});

export type F = typeof features;

export const columnHelper = createColumnHelper<F, dto.TextRow>();
