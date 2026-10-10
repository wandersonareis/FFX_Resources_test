import { resolveSegmentLabel } from '@/lib/ffx/display-names';
import type { EntryKind } from '@/lib/ffx/display-names';
import { columnHelper } from '../table-types';

/**
 * Coluna de nome — só existe em objects ("Nome") e lockit ("Tipo", com o
 * rótulo do segmento em vez do valor bruto).
 */
export function buildNameColumn(activeKind: EntryKind) {
  return columnHelper.accessor('name', {
    header: activeKind === 'lockit' ? 'Tipo' : 'Nome',
    cell: (info) =>
      activeKind === 'lockit' ? (
        <span className="text-muted-foreground">
          {resolveSegmentLabel(info.getValue())}
        </span>
      ) : (
        <span>{info.getValue()}</span>
      ),
  });
}
