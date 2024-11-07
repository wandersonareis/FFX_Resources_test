package internal

import tbstables "ffxresources/backend/formats/internal/tbs"

func decoderArgs() ([]string, string, error) {
	codeTable, err := tbstables.NewCharacterTable().GetFfx2CharacterTable()
	if err != nil {
		return nil, "", err
	}

	args := []string{
		"-e", "-t", codeTable,
	}

	return args, codeTable, nil
}
