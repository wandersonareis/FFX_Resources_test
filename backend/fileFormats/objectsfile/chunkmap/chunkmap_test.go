package chunkmap_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/objectsfile/chunkmap"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx2"
	testcommon "ffxresources/testData"
)

/*
Validação dos schemas chunkmap contra os binários reais — o fluxo novo é
AUTO-SUFICIENTE: nada aqui depende do KeyedStringFile legado para validar o
novo fluxo em si.

Para cada arquivo coberto por uma struct T:

  1. binary.Size(T) == individualLength lido do cabeçalho (V1 = 20 bytes,
     FFX; V2 = 32 bytes, FFX-2/LastMission) e == totalLength/count;
  2. round-trip Decode -> Encode byte-exato em TODOS os chunks;
  3. as chaves do objeto mapeado são exatamente as declaradas em wantKeys
     (snake_case do nome do campo C#, na ordem do arquivo);
  4. ToBytes sem edição devolve exatamente os bytes originais do chunk;
  5. ciclo de vida do arquivo inteiro: LoadFileFromBytes -> ToBytes reproduz
     o binário original byte-a-byte (header verbatim + chunks + string table
     reconstruída);
  6. export -> import identidade (importar o que foi exportado não muda
     nada) e o caminho dos builders (ExportFieldTexts) devolve o mesmo
     resultado — coberto por chunkmap/text_test.go.

Cruzamento com o fluxo legado (diagnóstico de migração, separado):
a comparação é por QUANTIDADE de campos de texto — independentemente de qual
lado tem mais ou menos — e pela IGUALDADE dos valores extraídos, posiçional.
Os nomes ("name", "help", ...) são apenas rótulos humanos: nunca participam
da comparação. Divergência de quantidade falha o teste e pede conferência no
editor hex (caso config_txt: o legado lia 2 campos e o binário tem 4 —
corrigido no LayoutSet; os demais devem casar).

Os binários vêm de build/bin/data (fonte primária — tem todos os arquivos de
FFX, FFX-2 e LastMission); testData/ e examples/ são fallbacks gitignored.
*/

const lang = common.DefaultLocalization

// TestMain prepara os charsets versionados (sem eles a resolução de texto
// aborta com encoding.ErrVersionNotFound). LastMiss reaproveita FFX-2.
func TestMain(m *testing.M) {
	for _, version := range []common.GameVersion{common.GameVersionFFX, common.GameVersionFFX2} {
		if err := ffxencoding.PrepareVersionCharsets(version); err != nil {
			fmt.Fprintf(os.Stderr, "preparar charsets %s: %v\n", version, err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}

// projectRoot sobe do diretório de test data até a raiz do módulo.
func projectRoot() string {
	return filepath.Dir(testcommon.GetTestDataRootDirectory())
}

// binaryCandidates lista, em ordem de preferência, os caminhos (relativos à
// raiz do projeto) em que o binário pode estar. build/bin/data tem TODOS os
// arquivos; testData/ e examples/ são os bins já usados antes (gitignored).
func binaryCandidates(version common.GameVersion, patternPath string) []string {
	game := "ffx2"
	if version == common.GameVersionFFX {
		game = "ffx"
	}
	suffix := "ffx_ps2/" + game + "/master/new_uspc/" + patternPath
	candidates := []string{"build/bin/data/" + suffix}
	switch version {
	case common.GameVersionFFX:
		candidates = append(candidates, "testData/FFX/binary/"+suffix)
	case common.GameVersionFFX2, common.GameVersionLastMiss:
		candidates = append(candidates, "testData/FFX-2/binary/"+suffix)
	}
	return append(candidates, "examples/"+patternPath)
}

func readBinary(t *testing.T, version common.GameVersion, patternPath string) (string, []byte) {
	t.Helper()
	candidates := binaryCandidates(version, patternPath)
	for _, rel := range candidates {
		data, err := os.ReadFile(filepath.Join(projectRoot(), rel))
		if err == nil {
			return rel, data
		}
	}
	t.Fatalf("binario %s (%s) nao encontrado; tentei:\n  %s",
		patternPath, version, strings.Join(candidates, "\n  "))
	return "", nil
}

// checkSchema roda as verificações auto-suficientes de um arquivo contra a
// struct T.
//
// wantNames contém os caminhos Go esperados dos campos textuais.
func checkSchema[T any](t *testing.T, patternPath string, version common.GameVersion, wantSize int, wantNames []string) {
	t.Helper()

	relPath, data := readBinary(t, version, patternPath)

	h, chunks, strtab, err := chunkmap.Split(data, version)
	if err != nil {
		t.Fatalf("%s: %v", relPath, err)
	}
	if got := chunkmap.StructSize[T](); got != wantSize {
		t.Fatalf("StructSize = %d, esperado %d", got, wantSize)
	}
	if h.IndividualLength != wantSize {
		t.Fatalf("individualLength do cabecalho = %d, esperado %d", h.IndividualLength, wantSize)
	}
	if h.TotalLength != h.Count()*wantSize {
		t.Fatalf("totalLength = %d != count(%d) * %d", h.TotalLength, h.Count(), wantSize)
	}

	fields, err := chunkmap.SegmentFields[T]()
	if err != nil {
		t.Fatalf("SegmentFields: %v", err)
	}
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.Name
	}
	if !slices.Equal(names, wantNames) {
		t.Fatalf("nomes = %v, esperado %v", names, wantNames)
	}

	// (2)+(4): round-trip binário e ToBytes byte-exato em TODOS os chunks
	for i := 0; i < h.ChunkCount(); i++ {
		raw := h.Chunk(chunks, i)

		chunk, err := chunkmap.Decode[T](raw)
		if err != nil {
			t.Fatalf("chunk %d: Decode: %v", i, err)
		}
		enc, err := chunk.Encode()
		if err != nil {
			t.Fatalf("chunk %d: Encode: %v", i, err)
		}
		if !bytes.Equal(enc, raw) {
			t.Fatalf("chunk %d: Encode nao reproduz os bytes originais", i)
		}

		obj, err := chunkmap.NewMappedTextObject[T](raw, strtab, h.IndividualLength, lang, version, patternPath)
		if err != nil {
			t.Fatalf("chunk %d: NewMappedTextObject: %v", i, err)
		}

		// export -> import identidade (sem nenhuma edicao)
		exported, err := obj.ExportText(lang)
		if err != nil {
			t.Fatalf("chunk %d: ExportText: %v", i, err)
		}
		if err := obj.ImportText(exported); err != nil {
			t.Fatalf("chunk %d: ImportText: %v", i, err)
		}
		out, err := obj.ToBytes(lang)
		if err != nil {
			t.Fatalf("chunk %d: ToBytes: %v", i, err)
		}
		if !bytes.Equal(out, raw) {
			t.Fatalf("chunk %d: ToBytes nao reproduz os bytes originais", i)
		}
	}

	// (5) ciclo de vida do arquivo inteiro, sem tocar no fluxo legado
	file, err := chunkmap.LoadFileFromBytes[T](version, patternPath, data)
	if err != nil {
		t.Fatalf("LoadFileFromBytes: %v", err)
	}
	if file.Count() != h.ChunkCount() {
		t.Fatalf("File.Count = %d, esperado %d", file.Count(), h.ChunkCount())
	}
	whole, err := file.ToBytes()
	if err != nil {
		t.Fatalf("File.ToBytes: %v", err)
	}
	if !bytes.Equal(whole, data) {
		t.Fatalf("File.ToBytes nao reproduz o arquivo original (%d vs %d bytes)",
			len(whole), len(data))
	}
}

// ---- caminhos Go esperados na ordem binária ----

var (
	// FFX — TextPair exposes its Standard and Simplified references.
	ffxCmdNames = []string{"Name.Standard", "Name.Simplified", "Desc.Standard", "Desc.Simplified"}
	ffxKeyItem  = []string{"Name.Standard", "Name.Simplified", "Help.Standard", "Help.Simplified"}
	ffxPanel    = []string{"Name.Standard", "Name.Simplified", "Help.Standard", "Help.Simplified"}
	ffxText2    = []string{"Help.Standard", "Help.Simplified"}
	ffxNameHelp = []string{"Command.Standard", "Command.Simplified", "Help.Standard", "Help.Simplified"}
	ffxMonStats = []string{"Name.Standard", "SensorText.Standard", "SensorText.Simplified", "ScanText.Standard", "ScanText.Simplified"}
	ffxPlyRom   = []string{"SwitchText.Standard", "SwitchText.Simplified", "ScanText.Standard", "ScanText.Simplified"}
	ffxPlySave  = []string{"Name.Standard"}

	// w_name.bin — Names (7 nomes) e NamesHira (7 leituras); um slot por
	// personagem, identificado pelos model_ids do próprio chunk (nunca pela
	// posição no array).
	ffxWeaponNames = []string{
		"Names[0].Standard", "Names[1].Standard", "Names[2].Standard", "Names[3].Standard",
		"Names[4].Standard", "Names[5].Standard", "Names[6].Standard",
		"NamesHira[0].Simplified", "NamesHira[1].Simplified", "NamesHira[2].Simplified",
		"NamesHira[3].Simplified", "NamesHira[4].Simplified", "NamesHira[5].Simplified",
		"NamesHira[6].Simplified",
	}

	// FFX-2 — TextRef fields export as Standard.
	ffx2NameHelp = []string{"Name.Standard", "Help.Standard"}
	ffx2Creature = []string{"Name.Standard", "Help.Standard", "CreatureData.Help.Standard"}
	ffx2Plate    = []string{
		"Name.Standard", "Help.Standard",
		"Messages[0].Standard", "Messages[1].Standard", "Messages[2].Standard", "Messages[3].Standard",
		"CreatureData.Help.Standard",
	}
	ffx2CmdText  = []string{"Command.Standard", "Help.Standard"}
	ffx2BtlTxt   = []string{"Help.Standard"}
	ffx2Oversoul = []string{"Name.Standard"}

	// Last Mission schemas.
	lm4    = []string{"Name.Standard", "Help.Standard", "Effect.Standard", "EffectDescription.Standard"}
	lm3    = []string{"Name.Standard", "Help.Standard", "Effect.Standard"}
	lmInfo = []string{"Name.Standard", "Help.Standard", "Information.Standard"}
	lm2    = []string{"Name.Standard", "Help.Standard"}
	lm1    = []string{"Name.Standard"}
)

// Caminhos (patternPath) dos binários de teste.
const (
	ffxKernelPath = "battle/kernel/"
	lastMissPath  = "lastmiss/kernel/"
)

// ---- FFX (V1) ----

func TestFFXCommandSchema(t *testing.T) {
	for _, file := range []string{"command", "item"} {
		t.Run(file, func(t *testing.T) {
			checkSchema[ffx.Command](t, ffxKernelPath+file+".bin", common.GameVersionFFX, 96, ffxCmdNames)
		})
	}
	t.Run("important", func(t *testing.T) {
		checkSchema[ffx.KeyItem](t, ffxKernelPath+"important.bin", common.GameVersionFFX, 20, ffxKeyItem)
	})
	t.Run("a_ability", func(t *testing.T) {
		checkSchema[ffx.AutoAbility](t, ffxKernelPath+"a_ability.bin", common.GameVersionFFX, 108, ffxCmdNames)
	})
	t.Run("panel", func(t *testing.T) {
		checkSchema[ffx.SphereGridNodeType](t, ffxKernelPath+"panel.bin", common.GameVersionFFX, 24, ffxPanel)
	})
	t.Run("sphere", func(t *testing.T) {
		checkSchema[ffx.Sphere](t, ffxKernelPath+"sphere.bin", common.GameVersionFFX, 16, ffxText2)
	})
	// monmagic{1|2}.bin = C# FFX.Command (0x5C), sem PCommandData: 4 campos
	// de texto (name + desc), todos confirmados no binário.
	for _, name := range []string{"monmagic1", "monmagic2"} {
		t.Run(name, func(t *testing.T) {
			checkSchema[ffx.MonMagic](t, ffxKernelPath+name+".bin", common.GameVersionFFX, 92, ffxCmdNames)
		})
	}
}

func TestFFXMonsterSchema(t *testing.T) {
	for _, name := range []string{"monster1", "monster2", "monster3"} {
		t.Run(name, func(t *testing.T) {
			checkSchema[ffx.MonStats](t, ffxKernelPath+name+".bin", common.GameVersionFFX, 128, ffxMonStats)
		})
	}
}

func TestFFXTextSchema(t *testing.T) {
	t.Run("btl_txt", func(t *testing.T) {
		checkSchema[ffx.HelpText](t, ffxKernelPath+"btl_txt.bin", common.GameVersionFFX, 8, ffxText2)
	})
	// arms/config/item/mmain/menu/summon/status/btlend/build/name/save_txt
	for _, name := range []string{
		"arms_txt", "btlend_txt", "build_txt", "config_txt", "item_txt",
		"menu_txt", "mmain_txt", "name_txt", "save_txt", "status_txt", "summon_txt",
	} {
		t.Run(name, func(t *testing.T) {
			checkSchema[ffx.NameHelpText](t, ffxKernelPath+name+".bin", common.GameVersionFFX, 16, ffxNameHelp)
		})
	}
}

func TestFFXPlayerSchema(t *testing.T) {
	t.Run("ply_rom", func(t *testing.T) {
		checkSchema[ffx.PlyRom](t, ffxKernelPath+"ply_rom.bin", common.GameVersionFFX, 44, ffxPlyRom)
	})
	t.Run("ply_save", func(t *testing.T) {
		checkSchema[ffx.PlySave](t, ffxKernelPath+"ply_save.bin", common.GameVersionFFX, 148, ffxPlySave)
	})
}

func TestFFXWeaponNameSchema(t *testing.T) {
	t.Run("w_name", func(t *testing.T) {
		checkSchema[ffx.WeaponName](t, ffxKernelPath+"w_name.bin", common.GameVersionFFX, 72, ffxWeaponNames)
	})
}

// TestFFXWeaponNameModelOwners valida o mapa model_id -> personagem contra o
// binário real: todo ID do arquivo resolve, cada registro tem um dono distinto
// por posição e a correspondência confere nos pontos conhecidos (0x4007 é a
// última arma do Tidus, 0x4042 o primeiro escudo, 0x4062 a última do Rikku).
func TestFFXWeaponNameModelOwners(t *testing.T) {
	relPath, data := readBinary(t, common.GameVersionFFX, ffxKernelPath+"w_name.bin")
	file, err := chunkmap.LoadFileFromBytes[ffx.WeaponName](common.GameVersionFFX, ffxKernelPath+"w_name.bin", data)
	if err != nil {
		t.Fatalf("LoadFileFromBytes: %v", err)
	}

	seen := make(map[ffx.WeaponModelKey]struct{})
	for i := 0; i < file.Count(); i++ {
		characters := make(map[string]struct{}, 7)
		for _, id := range file.Get(i).Value().ModelIDs {
			key := ffx.WeaponModelKeyOf(id)
			owner, ok := ffx.WeaponModelOwnerOf(key)
			if !ok {
				t.Fatalf("%s: chunk %d: model id %s sem dono no mapa", relPath, i, key)
			}
			if owner.Kind != ffx.WeaponKindWeapon && owner.Kind != ffx.WeaponKindShield {
				t.Fatalf("%s: chunk %d: model id %s com kind inesperado %q", relPath, i, key, owner.Kind)
			}
			seen[key] = struct{}{}
			characters[owner.Character] = struct{}{}
		}
		if len(characters) != 7 {
			t.Fatalf("%s: chunk %d: %d personagens distintos nos model_ids, esperado 7",
				relPath, i, len(characters))
		}
	}
	if len(seen) != 74 {
		t.Fatalf("%s: %d model ids distintos no arquivo, esperado 74", relPath, len(seen))
	}

	spot := []struct {
		key  ffx.WeaponModelKey
		want ffx.WeaponModelOwner
	}{
		{ffx.WeaponModelKey{0x07, 0x40}, ffx.WeaponModelOwner{Character: "Tidus", Kind: ffx.WeaponKindWeapon}},
		{ffx.WeaponModelKey{0x42, 0x40}, ffx.WeaponModelOwner{Character: "Tidus", Kind: ffx.WeaponKindShield}},
		{ffx.WeaponModelKey{0x0B, 0x40}, ffx.WeaponModelOwner{Character: "Yuna", Kind: ffx.WeaponKindWeapon}},
		{ffx.WeaponModelKey{0x4B, 0x40}, ffx.WeaponModelOwner{Character: "Yuna", Kind: ffx.WeaponKindShield}},
		{ffx.WeaponModelKey{0x62, 0x40}, ffx.WeaponModelOwner{Character: "Rikku", Kind: ffx.WeaponKindShield}},
	}
	for _, s := range spot {
		owner, ok := ffx.WeaponModelOwnerOf(s.key)
		if !ok {
			t.Fatalf("model id %s ausente no mapa", s.key)
		}
		if owner != s.want {
			t.Fatalf("model id %s = %+v, esperado %+v", s.key, owner, s.want)
		}
	}
}

// ---- FFX-2 (V2) ----

func TestFFX2BattleSchema(t *testing.T) {
	for _, c := range []struct {
		file string
		size int
		keys []string
		typ  string
	}{
		{"command", 140, ffx2NameHelp, "command"},   // PCommand = Command(134) + PCommandData(6)
		{"item", 160, ffx2NameHelp, "item"},         // Item = Command(134) + ItemData(26)
		{"monmagic", 136, ffx2NameHelp, "monmagic"}, // MCommand = Command(134) + MCommandData(2)
		{"important", 12, ffx2NameHelp, "important"},
		{"a_ability", 176, ffx2NameHelp, "a_ability"},
		{"accessory", 84, ffx2Creature, "accessory"},
		{"job", 228, ffx2Creature, "job"},
		{"plate", 128, ffx2Plate, "plate"},
		{"menu_txt", 8, ffx2CmdText, "menu_txt"},
		{"monster", 200, ffx2NameHelp, "monster"},
		{"monster2", 200, ffx2NameHelp, "monster2"},
		{"oversoul", 8, ffx2Oversoul, "oversoul"},
		{"btl_txt", 4, ffx2BtlTxt, "btl_txt"},
		{"btlend_txt", 8, ffx2CmdText, "btlend_txt"},
	} {
		c := c
		t.Run(c.file, func(t *testing.T) {
			switch c.typ {
			case "command":
				checkSchema[ffx2.PCommand](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "item":
				checkSchema[ffx2.Item](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "monmagic":
				checkSchema[ffx2.MCommand](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "important":
				checkSchema[ffx2.KeyItem](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "a_ability":
				checkSchema[ffx2.AutoAbility](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "accessory":
				checkSchema[ffx2.Accessory](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "job":
				checkSchema[ffx2.Job](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "plate":
				checkSchema[ffx2.Plate](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "menu_txt":
				checkSchema[ffx2.MenuTxt](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "monster":
				checkSchema[ffx2.Monster](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "monster2":
				checkSchema[ffx2.Monster2](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "oversoul":
				checkSchema[ffx2.Oversoul](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "btl_txt":
				checkSchema[ffx2.BtlTxt](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			case "btlend_txt":
				checkSchema[ffx2.BtlEndTxt](t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2, c.size, c.keys)
			}
		})
	}
}

// ---- Last Mission (V2) ----

func TestLastMissionSchema(t *testing.T) {
	for _, c := range []struct {
		file string
		size int
		keys []string
		typ  string
	}{
		{"lm_accesary", 52, lm4, "lm_accesary"},
		{"lm_command", 68, lmInfo, "lm_command"},
		{"lm_monmagic", 72, lmInfo, "lm_monmagic"},
		{"lm_monster", 128, lm2, "lm_monster"},
		{"lm_item", 60, lm3, "lm_item"},
		{"lm_dress", 40, lm3, "lm_dress"},
		{"lm_trap", 36, lm3, "lm_trap"},
		{"lm_mes", 8, lm2, "lm_mes"},
		{"lm_player", 60, lm2, "lm_player"},
		{"lm_warehouse", 60, lm2, "lm_warehouse"},
		{"lm_floorname", 4, lm1, "lm_floorname"},
	} {
		c := c
		t.Run(c.file, func(t *testing.T) {
			switch c.typ {
			case "lm_accesary":
				checkSchema[ffx2.LmAccesary](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss, c.size, c.keys)
			case "lm_command":
				checkSchema[ffx2.LmCommand](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss, c.size, c.keys)
			case "lm_monmagic":
				checkSchema[ffx2.LmMonMagic](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss, c.size, c.keys)
			case "lm_monster":
				checkSchema[ffx2.LmMonster](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss, c.size, c.keys)
			case "lm_item":
				checkSchema[ffx2.LmItem](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss, c.size, c.keys)
			case "lm_dress":
				checkSchema[ffx2.LmDress](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss, c.size, c.keys)
			case "lm_trap":
				checkSchema[ffx2.LmTrap](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss, c.size, c.keys)
			case "lm_mes":
				checkSchema[ffx2.LmMes](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss, c.size, c.keys)
			case "lm_player":
				checkSchema[ffx2.LmPlayer](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss, c.size, c.keys)
			case "lm_warehouse":
				checkSchema[ffx2.LmWarehouse](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss, c.size, c.keys)
			case "lm_floorname":
				checkSchema[ffx2.LmFloorName](t, lastMissPath+c.file+".bin", common.GameVersionLastMiss, c.size, c.keys)
			}
		})
	}
}
