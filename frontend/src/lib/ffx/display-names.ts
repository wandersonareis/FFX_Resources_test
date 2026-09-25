// Dicionário de exibição (frontend decide a apresentação).
// Backend envia apenas sumário canônico {id, key}; os labels PT-BR vivem aqui.
// Nomes dos GRUPOS de events (shortened) vivem em event-group-names.ts.

export type EntryKind = 'events' | 'objects' | 'macro' | 'lockit';

export const KIND_LABELS: Record<EntryKind, string> = {
  events: 'Eventos',
  objects: 'Sistema',
  macro: 'Dicionário',
  lockit: 'Loc Kit',
};

/** Tipo de armazenamento de um registro do lockit. */
export const LOCKIT_LABELS: Record<string, string> = {
  game: 'Jogo',
  utf8: 'UTF-8',
};

/** Collection id do lockit -> nome de exibição. */
export const LOCKIT_ENTRY_LABELS: Record<string, string> = {
  ffx_loc_kit_ps3: 'Kit de localização (FFX)',
  ffx2_loc_kit_ps3: 'Kit de localização (FFX-2)',
};

/** Collection id (basename sem extensão) -> nome de exibição. */
export const OBJECTS_LABELS: Record<string, string> = {
  important: 'Itens importantes',
  command: 'Comandos',
  item: 'Itens',
  a_ability: 'Habilidades',
  arms_txt: 'Armas',
  config_txt: 'Configurações',
  item_txt: 'Itens (textos)',
  mmain_txt: 'Menu principal',
  ply_rom: 'Quarto do jogador',
  btl_txt: 'Textos de batalha',
  btlend_txt: 'Fim de batalha',
  monmagic1: 'Magias de monstros 1',
  monmagic2: 'Magias de monstros 2',
  monster1: 'Monstros 1',
  monster2: 'Monstros 2',
  monster3: 'Monstros 3',
  build_txt: 'Construção',
  name_txt: 'Nomes',
  panel: 'Painel',
  sphere: 'Sphere Grid',
  save_txt: 'Save',
  status_txt: 'Status',
  summon_txt: 'Summons',
  w_name: 'Nomes de armas',
  accessory: 'Acessórios',
  job: 'Jobs',
  menu_txt: 'Menu',
  monmagic: 'Magias de monstros',
  monster: 'Monstros',
  oversoul: 'Oversoul',
  plate: 'Placas',
  ply_save: 'Save do jogador',
  lm_accesary: 'Acessórios (Last Mission)',
  lm_command: 'Comandos (Last Mission)',
  lm_dress: 'Dresspheres (Last Mission)',
  lm_floorname: 'Nomes de andares (Last Mission)',
  lm_item: 'Itens (Last Mission)',
  lm_mes: 'Mensagens (Last Mission)',
  lm_monmagic: 'Magias de monstros (Last Mission)',
  lm_monster: 'Monstros (Last Mission)',
  lm_player: 'Jogador (Last Mission)',
  lm_trap: 'Armadilhas (Last Mission)',
  lm_warehouse: 'Depósito (Last Mission)',
};

export function resolveEntryLabel(kind: EntryKind, id: string): string {
  if (kind === 'objects') {
    return OBJECTS_LABELS[id] ?? id;
  }
  if (kind === 'macro') {
    const m = /^chunk_(\d+)$/.exec(id);
    if (m) return `Bloco ${m[1]}`;
    return id;
  }
  if (kind === 'lockit') {
    return LOCKIT_ENTRY_LABELS[id] ?? id;
  }
  return id;
}

/** Rótulo do tipo (game/utf8) das rows do lockit. */
export function resolveSegmentLabel(name: string | undefined): string {
  if (!name) return '';
  return LOCKIT_LABELS[name] ?? name;
}
