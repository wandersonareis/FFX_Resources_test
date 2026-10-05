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

/**
 * Nomes legados por shortened (2 letras): grupos do FFX-2 que tinham nome mas
 * cujo fragmento exato é desconhecido. Só entra em cena quando o shortened não
 * tem nenhum fragmento nomeado.
 */
const LEGACY_SHORT_GROUPS: Record<string, string> = {
  au: 'Esfera do Auron',
  ca: 'Cartas',
  en: 'Final',
  ev: 'Evento',
  pa: 'Esfera da Dor',
  sa: 'Sabotagem',
};

export interface EventGroupResolution {
  /** Fragmento do nó alvo (onde os arquivos deste grupo moram). */
  target: string;
  /** Nome do local; vazio quando não nomeado (o nó exibe o próprio fragmento). */
  name: string;
}

/**
 * Resolve o nó de um fragmento de eventID:
 * - fragmento nomeado → nó próprio com o nome;
 * - sem nome → entra no PRIMEIRO fragmento nomeado do mesmo shortened
 *   (2 letras, ordem alfabética): azmm → azit (Al Bhed Home), bsyt → bsil
 *   (Besaid Island);
 * - nenhum nomeado no shortened → nome legado do shortened (Cartas...);
 * - senão → nó anônimo exibindo o próprio fragmento, igual ao nome da pasta
 *   no disco, para ser achado na busca da árvore de arquivos.
 */
export function resolveEventGroup(
  version: GameVersionId,
  fragment: string,
  siblings: string[],
): EventGroupResolution {
  const names = GROUPS_BY_VERSION[version] ?? {};
  const self = names[fragment];
  if (self) return { target: fragment, name: self };

  const namedSibling = siblings
    .filter((s) => s !== fragment && names[s])
    .sort()[0];
  if (namedSibling) return { target: namedSibling, name: names[namedSibling] };

  const legacy = LEGACY_SHORT_GROUPS[fragment.slice(0, 2)];
  if (legacy) return { target: fragment, name: legacy };

  return { target: fragment, name: '' };
}
