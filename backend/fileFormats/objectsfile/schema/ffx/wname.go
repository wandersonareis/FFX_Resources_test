package ffx

import (
	"fmt"

	"ffxresources/backend/models"
)

// w_name.bin — src/core/ffx/w_name.cs. individual = 0x48 (72 bytes).
//
// Cada chunk é uma entrada da tabela de nomes de equipamentos: 7 referências
// de nome (names), 7 da leitura em hiragana (names_hira), 7 model IDs
// (model_ids) e um ushort de reserva. O model ID identifica o personagem
// dono do nome via weaponModelOwners — a posição no array NUNCA é usada
// como identidade, nem no export nem no import.

// WeaponModelKey é o par de bytes do model_id EXATAMENTE como gravado no
// chunk (little-endian): o valor lido como 0x4007 está no arquivo como
// {0x07, 0x40}. A chave é o byte do dump hex, não o inteiro decodificado.
type WeaponModelKey [2]byte

// Kinds de equipamento resolvidos pelo mapa de model IDs.
const (
	WeaponKindWeapon = "weapon"
	WeaponKindShield = "shield"
)

// WeaponModelOwner é o dono resolvido de um model_id de w_name.bin.
type WeaponModelOwner struct {
	Character string
	Kind      string // WeaponKindWeapon | WeaponKindShield
}

// weaponModelOwners mapeia os bytes do model_id ao personagem dono. Tabela
// derivada do w_name.bin de referência (170 registros, 74 IDs únicos, cada
// par aparece em uma única posição dos arrays — zero conflitos), usando a
// ordem do fluxo antigo (T/Y/A/K/W/L/R) como correspondência real dos
// personagens. Um ID fora da tabela é erro: estenda o mapa, não presuma
// posição.
var weaponModelOwners = map[WeaponModelKey]WeaponModelOwner{
	// Tidus — armas {0x01,0x40}..{0x07,0x40} (0x4001..0x4007)
	{0x01, 0x40}: {"Tidus", WeaponKindWeapon},
	{0x02, 0x40}: {"Tidus", WeaponKindWeapon},
	{0x03, 0x40}: {"Tidus", WeaponKindWeapon},
	{0x04, 0x40}: {"Tidus", WeaponKindWeapon},
	{0x05, 0x40}: {"Tidus", WeaponKindWeapon},
	{0x06, 0x40}: {"Tidus", WeaponKindWeapon},
	{0x07, 0x40}: {"Tidus", WeaponKindWeapon},
	// Yuna — armas {0x0B,0x40}..{0x10,0x40} (0x400B..0x4010)
	{0x0B, 0x40}: {"Yuna", WeaponKindWeapon},
	{0x0C, 0x40}: {"Yuna", WeaponKindWeapon},
	{0x0D, 0x40}: {"Yuna", WeaponKindWeapon},
	{0x0E, 0x40}: {"Yuna", WeaponKindWeapon},
	{0x0F, 0x40}: {"Yuna", WeaponKindWeapon},
	{0x10, 0x40}: {"Yuna", WeaponKindWeapon},
	// Auron — armas {0x15,0x40}..{0x19,0x40} (0x4015..0x4019)
	{0x15, 0x40}: {"Auron", WeaponKindWeapon},
	{0x16, 0x40}: {"Auron", WeaponKindWeapon},
	{0x17, 0x40}: {"Auron", WeaponKindWeapon},
	{0x18, 0x40}: {"Auron", WeaponKindWeapon},
	{0x19, 0x40}: {"Auron", WeaponKindWeapon},
	// Kimahri — armas {0x1F,0x40}..{0x24,0x40} (0x401F..0x4024)
	{0x1F, 0x40}: {"Kimahri", WeaponKindWeapon},
	{0x20, 0x40}: {"Kimahri", WeaponKindWeapon},
	{0x21, 0x40}: {"Kimahri", WeaponKindWeapon},
	{0x22, 0x40}: {"Kimahri", WeaponKindWeapon},
	{0x23, 0x40}: {"Kimahri", WeaponKindWeapon},
	{0x24, 0x40}: {"Kimahri", WeaponKindWeapon},
	// Wakka — armas {0x29,0x40}..{0x2E,0x40} (0x4029..0x402E)
	{0x29, 0x40}: {"Wakka", WeaponKindWeapon},
	{0x2A, 0x40}: {"Wakka", WeaponKindWeapon},
	{0x2B, 0x40}: {"Wakka", WeaponKindWeapon},
	{0x2C, 0x40}: {"Wakka", WeaponKindWeapon},
	{0x2D, 0x40}: {"Wakka", WeaponKindWeapon},
	{0x2E, 0x40}: {"Wakka", WeaponKindWeapon},
	// Lulu — armas {0x33,0x40}..{0x38,0x40} (0x4033..0x4038)
	{0x33, 0x40}: {"Lulu", WeaponKindWeapon},
	{0x34, 0x40}: {"Lulu", WeaponKindWeapon},
	{0x35, 0x40}: {"Lulu", WeaponKindWeapon},
	{0x36, 0x40}: {"Lulu", WeaponKindWeapon},
	{0x37, 0x40}: {"Lulu", WeaponKindWeapon},
	{0x38, 0x40}: {"Lulu", WeaponKindWeapon},
	// Rikku — armas {0x3D,0x40}..{0x41,0x40} (0x403D..0x4041)
	{0x3D, 0x40}: {"Rikku", WeaponKindWeapon},
	{0x3E, 0x40}: {"Rikku", WeaponKindWeapon},
	{0x3F, 0x40}: {"Rikku", WeaponKindWeapon},
	{0x40, 0x40}: {"Rikku", WeaponKindWeapon},
	{0x41, 0x40}: {"Rikku", WeaponKindWeapon},
	// Tidus — escudos {0x42,0x40}..{0x46,0x40} (0x4042..0x4046)
	{0x42, 0x40}: {"Tidus", WeaponKindShield},
	{0x43, 0x40}: {"Tidus", WeaponKindShield},
	{0x44, 0x40}: {"Tidus", WeaponKindShield},
	{0x45, 0x40}: {"Tidus", WeaponKindShield},
	{0x46, 0x40}: {"Tidus", WeaponKindShield},
	// Yuna — escudos {0x47,0x40}..{0x4B,0x40} (0x4047..0x404B)
	{0x47, 0x40}: {"Yuna", WeaponKindShield},
	{0x48, 0x40}: {"Yuna", WeaponKindShield},
	{0x49, 0x40}: {"Yuna", WeaponKindShield},
	{0x4A, 0x40}: {"Yuna", WeaponKindShield},
	{0x4B, 0x40}: {"Yuna", WeaponKindShield},
	// Auron — escudos {0x4C,0x40}..{0x4F,0x40} (0x404C..0x404F)
	{0x4C, 0x40}: {"Auron", WeaponKindShield},
	{0x4D, 0x40}: {"Auron", WeaponKindShield},
	{0x4E, 0x40}: {"Auron", WeaponKindShield},
	{0x4F, 0x40}: {"Auron", WeaponKindShield},
	// Kimahri — escudos {0x50,0x40}..{0x54,0x40} (0x4050..0x4054)
	{0x50, 0x40}: {"Kimahri", WeaponKindShield},
	{0x51, 0x40}: {"Kimahri", WeaponKindShield},
	{0x52, 0x40}: {"Kimahri", WeaponKindShield},
	{0x53, 0x40}: {"Kimahri", WeaponKindShield},
	{0x54, 0x40}: {"Kimahri", WeaponKindShield},
	// Wakka — escudos {0x55,0x40}..{0x59,0x40} (0x4055..0x4059)
	{0x55, 0x40}: {"Wakka", WeaponKindShield},
	{0x56, 0x40}: {"Wakka", WeaponKindShield},
	{0x57, 0x40}: {"Wakka", WeaponKindShield},
	{0x58, 0x40}: {"Wakka", WeaponKindShield},
	{0x59, 0x40}: {"Wakka", WeaponKindShield},
	// Lulu — escudos {0x5A,0x40}..{0x5E,0x40} (0x405A..0x405E)
	{0x5A, 0x40}: {"Lulu", WeaponKindShield},
	{0x5B, 0x40}: {"Lulu", WeaponKindShield},
	{0x5C, 0x40}: {"Lulu", WeaponKindShield},
	{0x5D, 0x40}: {"Lulu", WeaponKindShield},
	{0x5E, 0x40}: {"Lulu", WeaponKindShield},
	// Rikku — escudos {0x5F,0x40}..{0x62,0x40} (0x405F..0x4062)
	{0x5F, 0x40}: {"Rikku", WeaponKindShield},
	{0x60, 0x40}: {"Rikku", WeaponKindShield},
	{0x61, 0x40}: {"Rikku", WeaponKindShield},
	{0x62, 0x40}: {"Rikku", WeaponKindShield},
}

// WeaponModelKeyOf devolve a chave de bytes (ordem do arquivo, LE) do ID
// decodificado pela struct.
func WeaponModelKeyOf(modelID uint16) WeaponModelKey {
	return WeaponModelKey{byte(modelID), byte(modelID >> 8)}
}

// WeaponModelOwnerOf resolve o dono do par de bytes gravado no chunk.
// ok=false significa model_id fora da tabela — erro, nunca fallback.
func WeaponModelOwnerOf(key WeaponModelKey) (WeaponModelOwner, bool) {
	owner, ok := weaponModelOwners[key]
	return owner, ok
}

// String rotula a chave em hex na ordem do arquivo ("07 40").
func (k WeaponModelKey) String() string {
	return fmt.Sprintf("%02X %02X", k[0], k[1])
}

// WeaponName (72 bytes) — battle/kernel/w_name.bin (individual = 0x48).
// Os rótulos textuais são Names[i].Standard e NamesHira[i].Simplified; o
// dono de cada slot é o personagem resolvido de ModelIDs[i] no mapa.
type WeaponName struct {
	Names       [7]models.TextRef       // names
	NamesHira   [7]models.SimplifiedRef // names_hira — leitura (Simplified)
	ModelIDs    [7]uint16               // model_ids — personagem dono de cada nome
	ReserveData uint16                  // _0x46 (reserva do layout)
}
