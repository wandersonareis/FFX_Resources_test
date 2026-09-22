// Nomes de exibição dos GRUPOS de events (prefixo eventID[:2]).
// O backend envia só o id canônico; o label PT-BR mora aqui (view decide).
//
// Os valores marcados com TODO são chutes/ainda não levantados: o fallback é
// exibir o próprio prefixo — troque pelo nome verdadeiro quando souber.
// Este arquivo é o único dono dos nomes de grupo, para não poluir o resto.

import type { GameVersionId } from './game-version';

/** Shortened do eventID: azit0000 → "az". */
export function shortenedOf(id: string): string {
  return id.slice(0, 2).toLowerCase();
}

/** FFX (ffx/event/obj_ps3/<az>/...). */
const FFX_GROUPS: Record<string, string> = {
  az: 'az', // TODO: nome real (azit*/azmm*)
  bi: 'Bikanel',
  bj: 'bj', // TODO: nome real (bjyt*)
  bl: 'Blitzball',
  bs: 'Besaid',
  bv: 'bv', // TODO: nome real (bvyt*/bvmm*)
  cd: 'Cid',
  cr: 'Créditos',
  dj: 'Djose',
  do: 'Dome',
  ga: 'Game Over',
  ge: 'ge', // TODO: nome real (genk*)
  gu: 'Guadosalam',
  hi: 'hi', // TODO: nome real (hiku*)
  ik: 'ik', // TODO: nome real (ikai*)
  is: 'is', // TODO: nome real (isho*)
  ka: 'Kamari',
  ki: 'ki', // TODO: nome real (kino*)
  kl: 'kl', // TODO: nome real (klyt*)
  lc: 'lc', // TODO: nome real (lchb*)
  lm: 'lm', // TODO: nome real (lmyt*) — NÃO é Last Mission
  lo: 'Demo',
  lu: 'Luca',
  ma: 'Macalania',
  mc: 'mc', // TODO: nome real (mcfr*/mcyt*)
  mi: "Mi'ihen",
  mm: 'mm', // TODO: nome real (mmmc*)
  ms: 'ms', // TODO: nome real (msmm*)
  mt: 'Mt. Gagazet',
  na: 'na', // TODO: nome real (nagi*)
  om: 'Omega',
  op: 'Abertura',
  pt: 'pt', // TODO: nome real (ptkl*)
  sa: 'Amostra',
  sc: 'Cenas',
  si: 'Sin',
  sl: 'sl', // TODO: nome real (slik*)
  ss: 'ss', // TODO: nome real (ssbt*)
  st: 'st', // TODO: nome real (stbv*)
  sw: 'sw', // TODO: nome real (swin*)
  sy: 'Sistema',
  te: 'Teste',
  zk: 'Zanarkand (ruínas)',
  zn: 'Zanarkand',
};

/**
 * FFX-2 (ffx2/event/obj_ps3/<az>/...). O grupo "lm" é o conteúdo da Last
 * Mission: fica oculto aqui na listagem e aparece na aba lastmiss.
 */
const FFX2_GROUPS: Record<string, string> = {
  ak: 'ak', // TODO: nome real (akagi*)
  au: 'Esfera do Auron',
  bi: 'Bikanel',
  bl: 'Blitzball',
  bs: 'Besaid',
  bv: 'bv', // TODO: nome real (bvyt*)
  ca: 'Cartas',
  cr: 'cr', // TODO: nome real (credits/crsm*)
  dj: 'Djose',
  dn: 'Dungeon', // TODO: confirmar (dnfr*)
  do: 'Dome',
  en: 'Final',
  ev: 'Evento',
  ga: 'Game Over',
  ge: 'ge', // TODO: nome real (genk*)
  gu: 'Guadosalam',
  hi: 'hi', // TODO: nome real (hiku*)
  ik: 'ik', // TODO: nome real (ikai*/ikaisphere)
  iw: 'iw', // TODO: nome real (iwatuto*)
  ka: 'Kamari',
  ki: 'ki', // TODO: nome real (kino*)
  kl: 'kl', // TODO: nome real (klyt*)
  lc: 'lc', // TODO: nome real (lchb*)
  lm: 'Last Mission',
  lu: 'Luca',
  ma: 'Macalania',
  mc: 'mc', // TODO: nome real (mcfr*)
  mi: "Mi'ihen",
  mo: 'mo', // TODO: nome real (monlist/movplay)
  mt: 'Mt. Gagazet',
  na: 'na', // TODO: nome real (nagi*)
  nu: 'nu', // TODO: nome real (nujisphere)
  op: 'Abertura',
  pa: 'Esfera da Dor',
  pt: 'pt', // TODO: nome real (ptkl*)
  sa: 'Sabotagem',
  sp: 'Esferas', // TODO: confirmar (spdn*/spheresel)
  st: 'st', // TODO: nome real (stbv*)
  su: 'su', // TODO: nome real (suka*)
  sy: 'sy', // TODO: nome real (syuin*)
  te: 'Teste',
  tu: 'tu', // TODO: nome real (tusinsel)
  we: 'we', // TODO: nome real (wegn*)
  yd: 'yd', // TODO: nome real (ydng*)
  yo: 'yo', // TODO: nome real (yougo*)
  zk: 'Zanarkand (ruínas)',
  zn: 'Zanarkand',
};

// lastmiss é expansão do ffx2 e lê a MESMA árvore de events.
const GROUPS_BY_VERSION: Record<GameVersionId, Record<string, string>> = {
  ffx: FFX_GROUPS,
  ffx2: FFX2_GROUPS,
  lastmiss: FFX2_GROUPS,
};

/** Label do grupo shortened da versão (fallback = o próprio prefixo). */
export function eventGroupLabel(version: GameVersionId, shortened: string): string {
  return GROUPS_BY_VERSION[version]?.[shortened] ?? shortened;
}
