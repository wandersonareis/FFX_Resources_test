package fileFormat

func dcpXpliterArgs() ([]string, error) {
	args := []string{
		"4", "-s", "-f", "-min", "0x40",
	}

	return args, nil
}

func dcpJoinerArgs() ([]string, error) {
	args := []string{
		"4", "-f", "-min", "0x40", "-i",
	}

	return args, nil
}