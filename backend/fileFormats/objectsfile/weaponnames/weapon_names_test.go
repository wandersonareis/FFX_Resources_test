package weaponnames_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/objectsfile/chunkmap"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx"
	"ffxresources/backend/fileFormats/objectsfile/weaponnames"
	testcommon "ffxresources/testData"
)

/*
Congelamento da política de w_name.bin contra o binário de referência.

w_name.bin é exclusivo do FFX e os personagens são fixos: a tabela
model_id -> dono (schema/ffx/wname.go) é ESTÁTICA. Estes testes provam os
dois sentidos da relação:

  1. arquivo -> tabela: todo par de bytes do binário tem dono (Validate);
  2. tabela -> arquivo: nenhuma entrada órfã — 74 chaves == 74 IDs
     observados, senão a tabela envelhece sem erro;
  3. whitelist: personagens e kinds restritos ao conjunto fechado;
  4. ordem de coluna do pristine: slot i resolve o personagem fixo i
     (assert de integridade do arquivo de referência — NÃO é restrição do
     fluxo, que aceita chunks reordenados com donos únicos);
  5. round-trip: Rows -> Apply sem edição -> ToBytes byte-exato.
*/

const lang = common.DefaultLocalization

func TestMain(m *testing.M) {
	if err := ffxencoding.PrepareVersionCharsets(common.GameVersionFFX); err != nil {
		fmt.Fprintf(os.Stderr, "preparar charsets: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func projectRoot() string {
	return filepath.Dir(testcommon.GetTestDataRootDirectory())
}

// readReference lê o w_name.bin de referência (build/bin/data é a fonte
// primária; testData/ e examples/ são fallbacks gitignored).
func readReference(t *testing.T) []byte {
	t.Helper()
	suffix := filepath.Join("ffx_ps2", "ffx", "master", "new_uspc", filepath.FromSlash(weaponnames.Pattern))
	candidates := []string{
		filepath.Join("build", "bin", "data", suffix),
		filepath.Join("testData", "FFX", "binary", suffix),
		filepath.Join("examples", filepath.FromSlash(weaponnames.Pattern)),
	}
	for _, rel := range candidates {
		data, err := os.ReadFile(filepath.Join(projectRoot(), rel))
		if err == nil {
			return data
		}
	}
	t.Fatalf("binário %s não encontrado; tentei:\n  %s", weaponnames.Pattern, strings.Join(candidates, "\n  "))
	return nil
}

func loadReference(t *testing.T, data []byte) *chunkmap.File[ffx.WeaponName] {
	t.Helper()
	file, err := chunkmap.LoadFileFromBytesForLanguage[ffx.WeaponName](
		common.GameVersionFFX, weaponnames.Pattern, lang, data)
	if err != nil {
		t.Fatalf("carregar: %v", err)
	}
	return file
}

// referenceTable devolve o conteúdo COMPLETO da tabela estática (varre o
// espaço de chaves de 2 bytes via WeaponModelOwnerOf).
func referenceTable() map[ffx.WeaponModelKey]ffx.WeaponModelOwner {
	table := make(map[ffx.WeaponModelKey]ffx.WeaponModelOwner)
	for v := 0; v <= 0xFFFF; v++ {
		key := ffx.WeaponModelKey{byte(v), byte(v >> 8)}
		if owner, ok := ffx.WeaponModelOwnerOf(key); ok {
			table[key] = owner
		}
	}
	return table
}

func TestWeaponNameTableMatchesReferenceBinary(t *testing.T) {
	file := loadReference(t, readReference(t))

	if err := weaponnames.Validate(file); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	table := referenceTable()
	if len(table) == 0 {
		t.Fatal("tabela de model ids vazia")
	}

	seen := make(map[ffx.WeaponModelKey]struct{})
	for chunk := 0; chunk < file.Count(); chunk++ {
		for _, id := range file.Get(chunk).Value().ModelIDs {
			seen[ffx.WeaponModelKeyOf(id)] = struct{}{}
		}
	}

	// tabela -> arquivo: nenhuma entrada órfã (a tabela não envelhece em
	// silêncio; qualquer ID a mais precisa aparecer no binário de referência).
	for key := range table {
		if _, ok := seen[key]; !ok {
			t.Fatalf("entrada órfã na tabela: %s — remova ou justifique", key)
		}
	}
	// arquivo -> tabela: Validate já prova, mas confirma o cardinal exato.
	if len(seen) != len(table) {
		t.Fatalf("cardinal divergente: %d IDs no arquivo vs %d entradas na tabela", len(seen), len(table))
	}
}

func TestWeaponNameTableWhitelist(t *testing.T) {
	allowedCharacters := map[string]bool{
		"Tidus": true, "Yuna": true, "Auron": true, "Kimahri": true,
		"Wakka": true, "Lulu": true, "Rikku": true,
	}
	allowedKinds := map[string]bool{
		ffx.WeaponKindWeapon: true,
		ffx.WeaponKindShield: true,
	}
	for key, owner := range referenceTable() {
		if !allowedCharacters[owner.Character] {
			t.Fatalf("%s: personagem fora do elenco fixo do FFX: %q", key, owner.Character)
		}
		if !allowedKinds[owner.Kind] {
			t.Fatalf("%s: kind inesperado %q", key, owner.Kind)
		}
	}
}

func TestWeaponNameColumnOrderOnReferenceBinary(t *testing.T) {
	file := loadReference(t, readReference(t))
	if err := weaponnames.AssertColumnOrder(file); err != nil {
		t.Fatal(err)
	}
}

func TestWeaponNameRowsRoundTripByteExact(t *testing.T) {
	data := readReference(t)
	file := loadReference(t, data)

	rows, err := weaponnames.Rows(file, map[string]*chunkmap.File[ffx.WeaponName]{lang: file})
	if err != nil {
		t.Fatalf("Rows: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("nenhuma row exportada")
	}
	for _, row := range rows {
		if row.Character == "" || row.Kind == "" {
			t.Fatalf("row %d/%s sem referência de dono: %+v", row.Index, row.Name, row)
		}
		if !strings.HasPrefix(row.Name, row.Character+".") {
			t.Fatalf("row %d: nome %q não rotula o dono %q", row.Index, row.Name, row.Character)
		}
	}

	// apply sem edição não pode tocar em byte nenhum.
	if err := weaponnames.Apply(file, rows); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	out, err := file.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes: %v", err)
	}
	if !bytes.Equal(out, data) {
		t.Fatalf("Rows->Apply->ToBytes não reproduz o original (%d vs %d bytes)", len(out), len(data))
	}
}
