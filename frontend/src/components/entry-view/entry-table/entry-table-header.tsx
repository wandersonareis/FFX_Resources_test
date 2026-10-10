import type { ReactTable } from '@tanstack/react-table';
import { TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { dto } from '@/wailsjs/go/models';
import type { F } from './table-types';

export function EntryTableHeader({
  table,
}: {
  table: ReactTable<F, dto.TextRow>;
}) {
  return (
    <TableHeader>
      {table.getHeaderGroups().map((headerGroup) => (
        <TableRow key={headerGroup.id}>
          {headerGroup.headers.map((header) => (
            <TableHead
              key={header.id}
              className={
                header.column.id === 'index' || header.column.id === 'seq'
                  ? 'w-14'
                  : undefined
              }
            >
              {header.isPlaceholder ? null : <table.FlexRender header={header} />}
            </TableHead>
          ))}
        </TableRow>
      ))}
    </TableHeader>
  );
}
