package reader

import (
    "encoding/json"
    "fmt"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "unicode/utf8"

    "ffxresources/backend/common"
    "ffxresources/backend/core/encoding"
)

const puaBase = 0xE000

func puaRune(code uint) rune { return rune(puaBase + int(code)) }

func tokenFor(code uint) string { return fmt.Sprintf("⟨%X⟩", code) }

var (
    rePUA   = regexp.MustCompile(`[\x{E000}-\x{F8FF}]`)
    reToken = regexp.MustCompile(`⟨([0-9A-Fa-f]+)⟩`)
)

func Detokenize(s string) string {
    return rePUA.ReplaceAllStringFunc(s, func(m string) string {
        r, _ := utf8.DecodeRuneInString(m)
        return tokenFor(uint(r - puaBase))
    })
}

func Tokenize(s string) (string, error) {
    var firstErr error
    out := reToken.ReplaceAllStringFunc(s, func(m string) string {
        sub := reToken.FindStringSubmatch(m)
        code, err := strconv.ParseUint(sub[1], 16, 32)
        if err != nil || code < 0x30 {
            if firstErr == nil {
                firstErr = fmt.Errorf("token inválido: %s", m)
            }
            return m
        }
        return string(puaRune(uint(code)))
    })
    return out, firstErr
}

type slotFix struct {
    code uint // código do jogo (0x30 + índice no arquivo)
    from rune // rune gravada no arquivo gerado (errada)
    to   rune // rune correta (confere com a font / tabela real)
}

var latinSlotFixes = []slotFix{
    {0xA7, 'ç', 'Ç'},        // gerador salvou ç minúsculo no slot do Ç maiúsculo
    {0x93, ' ', 'Œ'},        // font ímpar: ligadura OE entre ♪ e ”
    {0x97, ' ', 'œ'},        // font ímpar: ligadura oe entre ” e ↑
    {0xB5, 'Ù', puaRune(0xB5)}, // {xB5}: glifo não identificado → PUA fixa (round-trip)
    {0xB6, 'Ú', 'Ù'},        // gerador deslocou o bloco Ù Ú Û Ü ß em 1 slot
    {0xB7, 'Û', 'Ú'},
    {0xB8, 'Ü', 'Û'},
    {0xB9, 'ß', 'β'}, // {BETA}: se o glifo não for beta grego, troque por puaRune(0xB9)
}

func applySlotFixes(runes []rune, fixes []slotFix, charset string) {
    for _, f := range fixes {
        i := int(f.code) - 0x30
        if i < 0 || i >= len(runes) {
            continue
        }
        if runes[i] != f.from {
            if common.IsVerboseMode() {
                fmt.Printf("[charset %s] fix 0x%02X não aplicado: slot contém %q, esperado %q\n",
                    charset, f.code, runes[i], f.from)
            }
            continue
        }
        runes[i] = f.to
        if common.IsVerboseMode() {
            fmt.Printf("[charset %s] fix 0x%02X: %q → %q (U+%04X)\n", charset, f.code, f.from, f.to, f.to)
        }
    }
}

const maxSingleByteSlots = 0xFF - 0x30 + 1

func PrepareCharset(version common.GameVersion, charset string) error {
    version = version.Normalize()
    path := filepath.Join(
        common.GetPathRootForVersion(version),
        common.OriginalsFolder,
        common.GetEncodingPathForVersion(version, charset),
    )
    filePath, err := common.NewFileAccessor(path)
    if err != nil {
        return err
    }
    data := filePath.ReadBytes()
    if data == nil {
        return fmt.Errorf("charset %q: arquivo vazio ou ilegível: %s", charset, path)
    }

    str := strings.TrimPrefix(string(data), "\ufeff")
    str = strings.TrimRight(str, "\r\n")
    runes := []rune(str)

    applySlotFixes(runes, latinSlotFixes, charset)

    if replacements := loadCharReplacements(charset); len(replacements) > 0 {
        runes = applyReplacements(runes, replacements)
    }

    strict := len(runes) <= maxSingleByteSlots

    byteToChar, charToByte := buildMappings(runes, strict, charset)
    ffxencoding.SetCharMap(version, charset, byteToChar, charToByte)
    return nil
}

// PrepareAllCharsets carrega um charset para as duas versões (FFX e FFX-2).
// Erros de uma versão não bloqueiam a outra; retorna o último erro encontrado.
func PrepareAllCharsets(charset string) error {
    var lastErr error
    for _, v := range []common.GameVersion{common.GameVersionFFX, common.GameVersionFFX2} {
        if err := PrepareCharset(v, charset); err != nil {
            lastErr = err
        }
    }
    return lastErr
}


func buildMappings(runes []rune, strict bool, charset string) (map[uint]rune, map[rune]uint) {
    byteToChar := make(map[uint]rune)
    charToByte := make(map[rune]uint)
    dups := 0

    for i, r := range runes {
        idx := uint(0x30 + i)

        if r >= puaBase && r <= puaBase+0x8FF {
            fmt.Printf("[charset %s] AVISO: rune U+%04X no slot 0x%02X colide com a PUA gerada\n",
                charset, r, idx)
        }

        if prev, exists := charToByte[r]; exists {
            dups++
            if strict {
                byteToChar[idx] = puaRune(idx)
                charToByte[puaRune(idx)] = idx
                if common.IsVerboseMode() {
                    fmt.Printf("[charset %s] REPETIDO: %q em 0x%02X e 0x%02X → 0x%02X decodifica como %s\n",
                        charset, r, prev, idx, idx, tokenFor(idx))
                }
                continue
            }
            byteToChar[idx] = r
            if common.IsVerboseMode() {
                fmt.Printf("[charset %s] REPETIDO: %q (U+%04X) em 0x%02X e 0x%02X → encode usaria 0x%02X\n",
                    charset, r, r, prev, idx, prev)
            }
            continue
        }

        byteToChar[idx] = r
        charToByte[r] = idx
    }

    if common.IsVerboseMode() {
        fmt.Printf("[charset %s] resumo: %d slots, %d únicos, %d duplicados (estrito=%v)\n",
            charset, len(runes), len(charToByte), dups, strict)
    }
    return byteToChar, charToByte
}


func loadCharReplacements(charset string) map[rune]rune {
    path := filepath.Join(common.GetPathOriginalsRoot(), common.GetEncodingDir(), "char_replacements.json")
    resolvedFile, err := common.NewFileAccessor(path)
    if err != nil {
        fmt.Printf("Error resolving char replacements file: %v\n", err)
        return nil
    }
    data, err := common.ReadFile(resolvedFile.ResolvedPath)
    if err != nil {
        return nil
    }

    var config map[string]map[string]string
    if err := json.Unmarshal(data, &config); err != nil {
        return nil
    }
    charsetReplacements, ok := config[charset]
    if !ok {
        return nil
    }

    replacements := make(map[rune]rune)
    for oldChar, newChar := range charsetReplacements {
        oldCharRune := []rune(oldChar)
        newCharRune := []rune(newChar)
        if len(oldCharRune) == 1 && len(newCharRune) == 1 {
            replacements[oldCharRune[0]] = newCharRune[0]
        } else {
            fmt.Printf("Warning: Invalid replacement for charset %s: %s -> %s\n", charset, oldChar, newChar)
        }
    }
    return replacements
}

func applyReplacements(runes []rune, replacements map[rune]rune) []rune {
    if len(replacements) == 0 {
        return runes
    }

    newRunes := make([]rune, len(runes))
    copy(newRunes, runes)

    for i, r := range newRunes {
        if newChar, ok := replacements[r]; ok {
            newRunes[i] = newChar
        }
    }
    return newRunes
}