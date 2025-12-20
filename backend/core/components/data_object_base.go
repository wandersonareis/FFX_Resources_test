package components

/* type Stringer interface {
	String() string
} */

/* type DataObject interface {
	Stringer
	ToString(localization string) string
} */

type LocalizationSetter interface {
	SetLocalizations(other LocalizationSetter)
}

/* type DataObjectWithLocalizations interface {
	DataObject
	LocalizationSetter
} */

/* func ReadDataList[T ILocalizedTextObject](filename string, languageCode string, creator func([]byte, []byte, int, string) T) IList[T] {
	fileAccessor, err := common.NewFileAccessor(filename)
	if err != nil {
		if common.IsVerboseMode() {
			fmt.Printf("Erro ao acessar arquivo: %v\n", err)
		}
		return nil
	}

	if !fileAccessor.Exists {
		if common.IsVerboseMode() {
			fmt.Printf("Arquivo não existe: %s\n", filename)
		}
		return nil
	}

	data, err := os.ReadFile(fileAccessor.ResolvedPath)
	if err != nil {
		if common.IsVerboseMode() {
			fmt.Printf("Erro ao ler arquivo: %v\n", err)
		}
		return nil
	}

	if len(data) < 16 { // Minimum header size
		if common.IsVerboseMode() {
			fmt.Printf("Arquivo muito pequeno: %s\n", filename)
		}
		return nil
	}

	offset := 8

	minIndex := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	maxIndex := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Read individual length and total length
	individualLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	totalLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Skip 4 bytes
	offset += 4

	// Verify we have enough data
	if offset+totalLength > len(data) {
		if common.IsVerboseMode() {
			fmt.Printf("Dados insuficientes no arquivo: %s\n", filename)
		}
		return nil
	}

	// Read data bytes
	dataBytes := data[offset : offset+totalLength]
	offset += totalLength

	// Read string bytes (remaining data)
	stringBytes := data[offset:]

	// Create objects slice
	count := maxIndex - minIndex
	objectCount := count + 1
	objects := NewList[T](objectCount)

	for i := 0; i <= count; i++ {
		from := i * individualLength
		to := (i + 1) * individualLength

		if to > len(dataBytes) {
			break
		}
		objData := dataBytes[from:to]
		obj := creator(objData, stringBytes, individualLength, languageCode)
		objects.Add(obj)

		if common.IsVerboseMode() {
			offsetStr := fmt.Sprintf("%04X", (i*individualLength)+0x14)
			indexStr := fmt.Sprintf("%d", i+minIndex)
			fmt.Printf("%s (Offset %s) - %s\n", indexStr, offsetStr, obj.ToString(languageCode))
		}
	}

	return objects
} */

// Nova função compatível com IList usando interfaces
/* func ReadDataArrayWithIlist(filename string, languageCode string, creator func([]byte, []byte, int, string) ILocalizedTextObject) IList[ILocalizedTextObject] {
	list := ReadDataListWithIlist(filename, languageCode, creator)
	if list == nil {
		return nil
	}
	return list
} */

/* func ReadDataListWithIlist(filename string, languageCode string, creator func([]byte, []byte, int, string) ILocalizedTextObject) IList[ILocalizedTextObject] {
	fileAccessor, err := common.NewFileAccessor(filename)
	if err != nil {
		common.LogVerbose("Error accessing file: %v", err)
		return nil
	}

	if !fileAccessor.Exists {
		common.LogVerbose("File does not exist: %s", filename)
		return nil
	}

	data, err := os.ReadFile(fileAccessor.ResolvedPath)
	if err != nil {
		common.LogVerbose("Error reading file: %v", err)
		return nil
	}

	if len(data) < 16 { // Minimum header size
		common.LogVerbose("File too small: %s", filename)
		return nil
	}

	offset := 8

	minIndex := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	maxIndex := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Read individual length and total length
	individualLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	totalLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Skip 4 bytes
	offset += 4

	// Verify we have enough data
	if offset+totalLength > len(data) {
		common.LogVerbose("Insufficient data in file: %s", filename)
		return nil
	}

	// Read data bytes
	dataBytes := data[offset : offset+totalLength]
	offset += totalLength

	// Read string bytes (remaining data)
	stringBytes := data[offset:]

	// Create objects slice
	count := maxIndex - minIndex
	objectCount := count + 1
	objects := NewList[ILocalizedTextObject](objectCount)

	for i := 0; i <= count; i++ {
		from := i * individualLength
		to := (i + 1) * individualLength

		if to > len(dataBytes) {
			break
		}
		objData := dataBytes[from:to]
		obj := creator(objData, stringBytes, individualLength, languageCode)
		objects.Add(obj)

		if common.IsVerboseMode() {
			offsetStr := fmt.Sprintf("%04X", (i*individualLength)+0x14)
			indexStr := fmt.Sprintf("%d", i+minIndex)
			fmt.Printf("%s (Offset %s) - %s\n", indexStr, offsetStr, obj.ToString(languageCode))
		}
	}

	return objects
} */
