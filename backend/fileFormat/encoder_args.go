package fileFormat

func encoderArgs() ([]string, string, error) {
	codeTable, err := getCharacterTable()
	if err != nil {
		return nil, "", err
	}

	args := []string{
		"-i", "-t", codeTable,
	}

	return args, codeTable, nil
}