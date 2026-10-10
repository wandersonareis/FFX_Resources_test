/**
 * Tamanho legível para o desperdício do grupo de cópias. Uma casa decimal
 * a partir de KB (o byte fica inteiro) e sem falso "0.0 KB" para arquivos
 * minúsculos.
 */
export function formatBytes(value: number): string {
  if (!Number.isFinite(value) || value <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = value;
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit++;
  }
  return `${size >= 10 || unit === 0 ? size.toFixed(0) : size.toFixed(1)} ${units[unit]}`;
}
