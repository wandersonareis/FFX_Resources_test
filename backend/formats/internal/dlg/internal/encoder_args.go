package dlg_internal

import tbstables "ffxresources/backend/formats/internal/tbs"

func encoderArgs() ([]string, string, error) {
	codeTable, err := tbstables.NewCharacterTable().GetFfx2CharacterTable()
	if err != nil {
		return nil, "", err
	}

	args := []string{
		"-i", "-t", codeTable,
	}

	return args, codeTable, nil
}
