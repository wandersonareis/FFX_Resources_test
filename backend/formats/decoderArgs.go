package formats

func decoderArgs() ([]string, string, error) {
	codeTable, err := getFfx2CharacterTable()
	if err != nil {
		return nil, "", err
	}

	args := []string{
		"-e", "-t", codeTable,
	}

	return args, codeTable, nil
}
