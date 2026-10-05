package chunkmap_test

import (
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/fileFormats/objectsfile/chunkmap"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx"
	"ffxresources/backend/fileFormats/objectsfile/schema/ffx2"
	"ffxresources/backend/models"
)

/*
A guarda do ponteiro: o ref é DIREÇÃO, não fato.

O jogo sabe se existe texto olhando para a fronteira do bloco — não
tratando qualquer offset como uma frase. Um ref que cai no MEIO de outro
texto (MID) ou FORA da string table (OOB) não tem texto ali, e jamais pode
ser extraído como se tivesse.

Este arquivo fixa os três invariantes sobre binários reais (FFX, FFX-2 e
LastMission) e mede todos os outros ao lado:

  1. leitura — ref quebrado é classificado e expõe "" (nunca o sufixo do
     texto anterior, ex.: "s earned." em vez de "!");
  2. export — o ref quebrado não contribui linha nenhuma no idioma
     primário (sem linha não há diálogo, sem diálogo não há edição);
  3. save — ao editar UM campo legítimo, os segmentos dos refs quebrados
     permanecem byte-idênticos: o arquivo continua apontando exatamente
     para onde apontava.
*/

// pointered é o contrato observável de um ref lido — implementado por
// objectsfile.KeyedString (fluxo legado) e por chunkmap.TextContent. É por
// ele que a guarda é lida de fora do pacote.
type pointered interface {
	PointerStatus() objectsfile.PointerStatus
	GetOffset() models.Offset
	GetString() string
}

// slot identifica um campo de texto: chunk + posição do campo no layout.
type slot struct{ chunk, field int }

// slotKey identifica um campo pelo par (chunk, chave) — o que o export usa.
type slotKey struct {
	chunk int
	key   string
}

// slotState é o estado de um campo no idioma primário.
type slotState struct {
	key    string
	status objectsfile.PointerStatus
	offset models.Offset
	text   string
}

// snapshotOf captura o estado de todos os campos textuais de todos os chunks.
func snapshotOf[T any](f *chunkmap.File[T]) map[slot]slotState {
	out := make(map[slot]slotState)
	for ci, obj := range f.Objects {
		for fi, key := range obj.OrderedFieldKeys() {
			st := slotState{key: key}
			if seg := obj.GetKeyedString(key); seg != nil {
				if p, ok := seg.GetLocalizedContent(lang).(pointered); ok {
					st.status = p.PointerStatus()
					st.offset = p.GetOffset()
					st.text = p.GetString()
				}
			}
			out[slot{ci, fi}] = st
		}
	}
	return out
}

// countStatuses separa os campos por status do seu ref.
func countStatuses(before map[slot]slotState) (ok, empty, mid, oob, created int) {
	for _, st := range before {
		switch st.status {
		case objectsfile.PointerEmpty:
			empty++
		case objectsfile.PointerMid:
			mid++
		case objectsfile.PointerOOB:
			oob++
		case objectsfile.PointerCreated:
			created++
		default:
			ok++
		}
	}
	return
}

// reportPointer loga a medição de um arquivo (Fase 0: medir tudo antes de
// ligar a guarda onde há problema).
func reportPointer(t *testing.T, pattern string, before map[slot]slotState) {
	t.Helper()
	ok, empty, mid, oob, created := countStatuses(before)
	t.Logf("%-42s campos=%-5d ok=%-5d empty=%-5d mid=%-5d oob=%-5d created=%d",
		pattern, len(before), ok, empty, mid, oob, created)
}

// exportedPrimary devolve o texto do idioma primário exportado por
// (chunk, chave): o export é o único caminho que vira linha de texto.
func exportedPrimary[T any](f *chunkmap.File[T]) map[slotKey]string {
	out := make(map[slotKey]string)
	texts, err := f.ExportText(lang)
	if err != nil {
		return out
	}
	for _, ft := range texts {
		for _, fld := range ft.Fields {
			out[slotKey{chunk: ft.Index, key: fld.Key}] = fld.Texts[lang]
		}
	}
	return out
}

// checkGuardedReferences roda as verificações 1 e 2. wantBroken=false mede
// um arquivo que a medição diz estar limpo.
func checkGuardedReferences[T any](t *testing.T, pattern string, version common.GameVersion, wantBroken bool) {
	t.Helper()

	_, data := readBinary(t, version, pattern)
	f, err := chunkmap.LoadFileFromBytes[T](version, pattern, data)
	if err != nil {
		t.Fatalf("%s: %v", pattern, err)
	}
	before := snapshotOf(f)
	reportPointer(t, pattern, before)

	_, _, mid, oob, _ := countStatuses(before)
	broken := mid + oob
	if wantBroken && broken == 0 {
		t.Errorf("%s: nenhum ref quebrado (mid=%d oob=%d) — a medição mudou", pattern, mid, oob)
	}
	if !wantBroken && broken > 0 {
		t.Errorf("%s: %d ref(s) quebrado(s) (mid=%d oob=%d) em arquivo medido como limpo", pattern, broken, mid, oob)
	}
	if broken == 0 {
		return
	}

	exported := exportedPrimary(f)
	for k, st := range before {
		if !st.status.IsBroken() {
			continue
		}
		if st.text != "" {
			t.Errorf("%s: chunk %d campo %q [%s @ %d]: ref quebrado leu texto %q",
				pattern, k.chunk, st.key, st.status, st.offset, st.text)
		}
		if got := exported[slotKey{chunk: k.chunk, key: st.key}]; got != "" {
			t.Errorf("%s: chunk %d campo %q [%s @ %d]: ref quebrado virou linha %q no export",
				pattern, k.chunk, st.key, st.status, st.offset, got)
		}
	}
}

// checkSaveKeepsSegments edita UM campo legítimo e confere que nada além
// dele mudou de texto — e que todo o resto andou JUNTO quando o arquivo
// cresceu.
//
// O rebuild faz splice do texto editado no próprio bloco: o deslocamento é
// GLOBAL. Os segmentos que vêm depois do bloco editado andam pelo mesmo
// delta (incluindo os quebrados — eles RECEBEM deslocamento sem CAUSAR),
// e os de antes ficam parados. O que os quebrados não podem é virar
// ponteiro válido.
func checkSaveKeepsSegments[T any](t *testing.T, pattern string, version common.GameVersion, wantBroken bool) {
	t.Helper()

	_, data := readBinary(t, version, pattern)
	f, err := chunkmap.LoadFileFromBytes[T](version, pattern, data)
	if err != nil {
		t.Fatalf("%s: %v", pattern, err)
	}
	before := snapshotOf(f)
	_ = data

	// Escolhe o campo legítimo mais curto (ref OK com texto) e o dobra:
	// garante a edição E que o texto resultante é codificável.
	var (
		edited  slot
		newText string
		found   bool
	)
	for ci := range f.Objects {
		for fi, key := range f.Objects[ci].OrderedFieldKeys() {
			st := before[slot{ci, fi}]
			if st.status != objectsfile.PointerOK || st.text == "" {
				continue
			}
			if found && len(newText) <= len(st.text) {
				continue
			}
			edited, newText, found = slot{ci, fi}, st.text+st.text, true
			_ = key
		}
	}
	if !found {
		t.Fatalf("%s: nenhum campo legítimo para editar", pattern)
	}
	if err := f.Objects[edited.chunk].ImportText([]objectsfile.FieldText{
		{Key: before[edited].key, Texts: map[string]string{lang: newText}},
	}); err != nil {
		t.Fatalf("%s: ImportText: %v", pattern, err)
	}
	if !f.Edited() {
		t.Fatalf("%s: a edição não marcou o arquivo como editado", pattern)
	}

	out, err := f.ToBytes()
	if err != nil {
		t.Fatalf("%s: ToBytes: %v", pattern, err)
	}
	reloaded, err := chunkmap.LoadFileFromBytes[T](version, pattern, out)
	if err != nil {
		t.Fatalf("%s: reload do save: %v", pattern, err)
	}
	after := snapshotOf(reloaded)

	// Deslocamento global: todo segmento INTOCADO anda pelo mesmo delta.
	// Os que vêm depois do bloco editado andam juntos pelo crescimento; os
	// de antes ficam parados. Os quebrados seguem a mesma regra — recebem
	// deslocamento sem causar deslocamento.
	editedOld := int(before[edited].offset)
	fullDelta, deltaDefinido := 0, false
	for k, old := range before {
		if k == edited {
			continue
		}
		cur, okAfter := after[k]
		if !okAfter {
			continue
		}
		got := int(cur.offset) - int(old.offset)
		if int(old.offset) <= editedOld {
			if got != 0 {
				t.Errorf("%s: chunk %d campo %q: offset %d -> %d, mas vem antes do bloco editado",
					pattern, k.chunk, old.key, old.offset, cur.offset)
			}
			continue
		}
		if !deltaDefinido {
			fullDelta, deltaDefinido = got, true
			continue
		}
		if got != fullDelta {
			t.Errorf("%s: chunk %d campo %q: offset %d -> %d (delta %+d), mas o deslocamento global é %+d",
				pattern, k.chunk, old.key, old.offset, cur.offset, got, fullDelta)
		}
	}

	keptSegments := 0
	for k, old := range before {
		cur, okAfter := after[k]
		if !okAfter {
			t.Errorf("%s: chunk %d campo %q sumiu do arquivo salvo", pattern, k.chunk, old.key)
			continue
		}
		if k == edited {
			if cur.text != newText {
				t.Errorf("%s: chunk %d campo %q: edição não gravou (esperado %q, lido %q)",
					pattern, k.chunk, old.key, newText, cur.text)
			}
			continue
		}
		if cur.text != old.text {
			t.Errorf("%s: chunk %d campo %q: texto mudou de %q para %q sem edição",
				pattern, k.chunk, old.key, old.text, cur.text)
		}
		if !old.status.IsBroken() {
			continue
		}
		keptSegments++
		// O que o jogo lê é "há texto?". OOB pode virar empty quando a
		// tabela cresce (os dois significam SEM texto); o que não pode
		// acontecer é virar ponteiro válido.
		if cur.status == objectsfile.PointerOK || cur.text != "" {
			t.Errorf("%s: chunk %d campo %q: segmento quebrado virou ponteiro válido (%s @ %d, texto %q)",
				pattern, k.chunk, old.key, cur.status, cur.offset, cur.text)
		}
	}
	if wantBroken && keptSegments == 0 {
		t.Errorf("%s: nenhum segmento quebrado sobreviveu ao save", pattern)
	}
}

// guardOf amarra um tipo concreto ao ciclo completo de guarda.
func guardOf[T any](wantBroken bool) func(t *testing.T, pattern string, version common.GameVersion) {
	return func(t *testing.T, pattern string, version common.GameVersion) {
		checkGuardedReferences[T](t, pattern, version, wantBroken)
		checkSaveKeepsSegments[T](t, pattern, version, wantBroken)
	}
}

// reportOf amarra um tipo concreto à medição pura (sem expectativa).
func reportOf[T any](t *testing.T, pattern string, version common.GameVersion) {
	t.Helper()
	_, data := readBinary(t, version, pattern)
	f, err := chunkmap.LoadFileFromBytes[T](version, pattern, data)
	if err != nil {
		t.Errorf("%s: %v", pattern, err)
		return
	}
	reportPointer(t, pattern, snapshotOf(f))
}

// ---- FFX (V1): guarda ligada nos arquivos medidos com defeito ----
//
// a_ability.bin não entra aqui: o binário declara 108 bytes por chunk e a
// struct FFX correspondente tem 96 — o LoadFileFromBytes recusa. Ele é
// coberto pelo legacy_crosscheck (que não valida tamanho de struct).

func TestPointerGuardFFX(t *testing.T) {
	for _, c := range []struct {
		file string
		run  func(t *testing.T, pattern string, version common.GameVersion)
	}{
		// Medidos com MID/OOB: exatamente onde a guarda precisa ligar.
		{"build_txt", guardOf[ffx.NameHelpText](true)},
		{"btlend_txt", guardOf[ffx.NameHelpText](true)},
		{"name_txt", guardOf[ffx.NameHelpText](true)},
		{"save_txt", guardOf[ffx.NameHelpText](true)},

		// Medidos limpos: a guarda tem que continuar DESLIGADA (o rebuild
		// dedup de sempre roda, como antes da mudança).
		{"command", guardOf[ffx.Command](false)},
		{"item", guardOf[ffx.Command](false)},
		{"important", guardOf[ffx.KeyItem](false)},
		{"panel", guardOf[ffx.SphereGridNodeType](false)},
		{"sphere", guardOf[ffx.Sphere](false)},
		{"btl_txt", guardOf[ffx.HelpText](false)},
		{"monmagic1", guardOf[ffx.MonMagic](false)},
		{"monmagic2", guardOf[ffx.MonMagic](false)},
		{"monster1", guardOf[ffx.MonStats](false)},
		{"monster2", guardOf[ffx.MonStats](false)},
		{"monster3", guardOf[ffx.MonStats](false)},
		{"menu_txt", guardOf[ffx.NameHelpText](false)},
		{"config_txt", guardOf[ffx.NameHelpText](false)},
		{"item_txt", guardOf[ffx.NameHelpText](false)},
		{"arms_txt", guardOf[ffx.NameHelpText](false)},
		{"status_txt", guardOf[ffx.NameHelpText](false)},
		{"summon_txt", guardOf[ffx.NameHelpText](false)},
		{"mmain_txt", guardOf[ffx.NameHelpText](false)},
		{"ply_rom", guardOf[ffx.PlyRom](false)},
		{"ply_save", guardOf[ffx.PlySave](false)},
	} {
		c := c
		t.Run(c.file, func(t *testing.T) {
			c.run(t, ffxKernelPath+c.file+".bin", common.GameVersionFFX)
		})
	}
}

// ---- FFX-2 (V2): Fase 0, medição ----

func TestPointerReportFFX2(t *testing.T) {
	for _, c := range []struct {
		file string
		run  func(t *testing.T, pattern string, version common.GameVersion)
	}{
		{"command", reportOf[ffx2.PCommand]},
		{"item", reportOf[ffx2.Item]},
		{"monmagic", reportOf[ffx2.MCommand]},
		{"important", reportOf[ffx2.KeyItem]},
		{"a_ability", reportOf[ffx2.AutoAbility]},
		{"accessory", reportOf[ffx2.Accessory]},
		{"job", reportOf[ffx2.Job]},
		{"plate", reportOf[ffx2.Plate]},
		{"menu_txt", reportOf[ffx2.MenuTxt]},
		{"monster", reportOf[ffx2.Monster]},
		{"monster2", reportOf[ffx2.Monster2]},
		{"oversoul", reportOf[ffx2.Oversoul]},
		{"btl_txt", reportOf[ffx2.BtlTxt]},
		{"btlend_txt", reportOf[ffx2.BtlEndTxt]},
	} {
		c := c
		t.Run(c.file, func(t *testing.T) {
			c.run(t, ffxKernelPath+c.file+".bin", common.GameVersionFFX2)
		})
	}
}

// ---- LastMission (V2): Fase 0, medição ----

func TestPointerReportLastMiss(t *testing.T) {
	for _, c := range []struct {
		file string
		run  func(t *testing.T, pattern string, version common.GameVersion)
	}{
		{"lm_command", reportOf[ffx2.LmCommand]},
		{"lm_item", reportOf[ffx2.LmItem]},
		{"lm_monmagic", reportOf[ffx2.LmMonMagic]},
		{"lm_monster", reportOf[ffx2.LmMonster]},
		{"lm_accesary", reportOf[ffx2.LmAccesary]},
		{"lm_dress", reportOf[ffx2.LmDress]},
		{"lm_trap", reportOf[ffx2.LmTrap]},
		{"lm_mes", reportOf[ffx2.LmMes]},
		{"lm_player", reportOf[ffx2.LmPlayer]},
		{"lm_warehouse", reportOf[ffx2.LmWarehouse]},
		{"lm_floorname", reportOf[ffx2.LmFloorName]},
	} {
		c := c
		t.Run(c.file, func(t *testing.T) {
			c.run(t, lastMissPath+c.file+".bin", common.GameVersionLastMiss)
		})
	}
}
