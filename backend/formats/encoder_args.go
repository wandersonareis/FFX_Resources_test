package formats

func encoderArgs() ([]string, string, error) {
	codeTable, err := getFfx2CharacterTable()
	if err != nil {
		return nil, "", err
	}

	args := []string{
		"-i", "-t", codeTable,
	}

	return args, codeTable, nil
}
