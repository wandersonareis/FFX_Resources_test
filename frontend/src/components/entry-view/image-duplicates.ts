/**
 * Rótulo de estado de uma cópia, relativo à textura de referência:
 *
 *	idêntica  → ainda sincronizada com o que está na tela;
 *	importada → editada em separado (divergente — sobrescrita por replicar);
 *	pristine  → nunca tocada (mas diferente da imagem de referência).
 */
export function copyState(d: { identical: boolean; modded: boolean }): string {
  if (d.identical) return 'idêntica';
  return d.modded ? 'importada' : 'pristine';
}
