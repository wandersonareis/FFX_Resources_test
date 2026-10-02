// Nomes de exibição dos GRUPOS de events (fragmento eventID[:4]).
// O backend envia só o id canônico; o label mora aqui (view decide).
//
// Chaves = os 4 primeiros caracteres do eventID (o mesmo nome da pasta em
// event/obj_ps3/<xx>/<fragmento>/<fragmento>####/). Fragmentos sem nome na
// tabela caem no fallback: o próprio fragmento é exibido (ex.: azmm, bsyt,
// gemm, isho, kamm...). Troque pelo nome verdadeiro quando souber.
// Este arquivo é o único dono dos nomes de grupo, para não poluir o resto.

import type { GameVersionId } from './game-version';

/** Fragmento do eventID: azit0000 → "azit" (mesmo nome da pasta no disco). */
export function shortenedOf(id: string): string {
  return id.slice(0, 4).toLowerCase();
}

/** FFX (ffx/event/obj_ps3/<fragmento>/<id>/<id>.bin). */
const FFX_GROUPS: Record<string, string> = {
  // Locais do jogo (nome por fragmento da pasta).
  azit: 'Al Bhed Home',
  bika: 'Bikanel Desert',
  bjyt: 'Baaj Temple',
  bltz: 'Blitzball Stadium',
  bsil: 'Besaid Island',
  bsmm: 'Besaid Beach (Flashback)',
  bsvr: 'Besaid Village',
  bvyt: 'Besaid Temple',
  cdsp: 'Al Bhed Boat & Underwater Ruins',
  djyt: 'Djose Temple',
  dome: 'Zanarkand Dome',
  dream: 'Unknown',
  genk: 'Moonflow',
  grid: 'Sphere Grid Plane',
  guad: 'Guadosalam',
  hiku: 'Airship & World Map',
  ikai: 'Farplane',
  kami: 'Thunder Plains',
  kino: 'Mushroom Rock',
  klyt: 'Kilika Woods & Temple',
  lchb: 'Luca',
  lmyt: 'Remiem Temple',
  luca: 'Luca Square & Pre-Rendered Backgrounds',
  maca: 'Lake Macalania',
  mcfr: 'Macalania Forest',
  mcyt: 'Macalania Temple',
  mihn: "Mi'hen Highroad",
  mmmc: 'Unknown',
  msmm: 'Via Purifico (Maze)',
  mtgz: 'Mt. Gagazet, Caves, Upper Zanarkand',
  nagi: 'Calm Lands & Cavern of the Stolen Fayth',
  omeg: 'Omega Ruins',
  ptkl: 'Kilika Town',
  sins: 'Inside Sin',
  slik: 'SS Liki',
  ssbt: 'Airship Model',
  stbv: 'Bevelle Highbridge, Via Purifico (Sewer), FFX-2 Map',
  swin: 'SS Winno',
  titl: 'Main Menu',
  zkrn: 'Zanarkand Ruins',
  znkd: 'Dream Zanarkand',
  zzzz: 'Unknown',

  // Pastas utilitárias (sufixo não é código de local).
  cred: 'Créditos',
  game: 'Game Over',
  loop: 'Demo',
  open: 'Abertura',
  samp: 'Amostra',
  scen: 'Cenas',
  sysf: 'Sistema',
  test: 'Teste',
};

/**
 * FFX-2 (ffx2/event/obj_ps3/<fragmento>/...). Usa os MESMOS fragmentos do FFX
 * (hiku, stbv, ikai, klyt...) + os específicos abaixo. O grupo "lmhi" é o
 * conteúdo da Last Mission: fica oculto na listagem do ffx2 e aparece na aba
 * lastmiss.
 */
const FFX2_EXTRA: Record<string, string> = {
  dnfr: 'Dungeon', // TODO: confirmar (dnfr####)
  lmhi: 'Last Mission', // lmhiku####
  spdn: 'Esferas', // TODO: confirmar (spdn*/spheresel)
  sphe: 'Esferas', // TODO: confirmar
};
const FFX2_GROUPS: Record<string, string> = { ...FFX_GROUPS, ...FFX2_EXTRA };

// lastmiss é expansão do ffx2 e lê a MESMA árvore de events.
const GROUPS_BY_VERSION: Record<GameVersionId, Record<string, string>> = {
  ffx: FFX_GROUPS,
  ffx2: FFX2_GROUPS,
  lastmiss: FFX2_GROUPS,
};

/** Label do grupo (fragmento do eventID) da versão; fallback = o próprio fragmento. */
export function eventGroupLabel(version: GameVersionId, shortened: string): string {
  return GROUPS_BY_VERSION[version]?.[shortened] ?? shortened;
}
