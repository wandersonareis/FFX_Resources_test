export const GAME_VERSIONS = [
  { id: 'ffx', label: 'FFX' },
  { id: 'ffx2', label: 'FFX-2' },
  { id: 'lastmiss', label: 'Last Mission' },
] as const;

export type GameVersionId = (typeof GAME_VERSIONS)[number]['id'];

export function isGameVersionId(value: unknown): value is GameVersionId {
  return (
    typeof value === 'string' &&
    (GAME_VERSIONS as readonly { id: string }[]).some((v) => v.id === value)
  );
}
