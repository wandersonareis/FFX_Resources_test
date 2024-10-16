package formats

func decoderArgs() ([]string, string, error) {
	codeTable, err := getCharacterTable()
	if err != nil {
		return nil, "", err
	}

	args := []string{
		"-e", "-t", codeTable,
	}

	return args, codeTable, nil
}
