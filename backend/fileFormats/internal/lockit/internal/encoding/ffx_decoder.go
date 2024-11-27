package lockitencoding

import (
	//"ffxresources/backend/fileFormats/util"
	//"ffxresources/backend/interactions"
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/lib"
	"fmt"
)

type LockitDecoder struct {}

func NewDecoder() *LockitDecoder {
	return &LockitDecoder{}
}

func (ld *LockitDecoder) LockitDecoderLoc(sourceFile, targetFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	encodingFile := encoding.GetFFXTextLockitLocalizationEncoding()

	executable, err := encoding.GetLockitFileHandler().FetchLockitHandler()
	if err != nil {
		return err
	}

	if err := ld.decoder(executable, sourceFile, targetFile, encodingFile); err != nil {
		return err
	}

	return nil
}

func (ld *LockitDecoder) LockitDecoderFfx(sourceFile, targetFile string, encoding ffxencoding.IFFXTextLockitEncoding) error {
	encodingFile := encoding.GetFFXTextLockitEncoding()

	executable, err := encoding.GetLockitFileHandler().FetchLockitHandler()
	if err != nil {
		return err
	}

	if err := ld.decoder(executable, sourceFile, targetFile, encodingFile); err != nil {
		return err
	}

	return nil
}

func (d *LockitDecoder) decoder(executable, sourceFile, targetFile string, encoding string) error {
	if !common.IsFileExists(sourceFile) {
		return fmt.Errorf("source file does not exist")
	}	

	if !common.IsFileExists(encoding) {
		return fmt.Errorf("encoding file does not exist")
	}

	args := []string{"-t", encoding, sourceFile, targetFile}

	if err := lib.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}

/* func LockitDecoderFfx(lockitFileInfo interactions.IGameDataInfo) error {
	characterTable := util.NewCharacterTable()
	
	codeTable, err := characterTable.GetCharacterOnlyTable()
	if err != nil {
		return err
	}

	defer characterTable.Dispose(codeTable)

	handler := lib.NewLockitHandler()
	defer handler.dispose()

	executable, err := handler.getLockitFileHandler()
	if err != nil {
		return err
	}

	return Decoder(lockitFileInfo, codeTable)
} */

/* func LockitDecoderLoc(lockitFileInfo interactions.IGameDataInfo) error {
	characterTable := util.NewCharacterTable()
	
	codeTable, err := characterTable.GetCharacterLocTable()
	if err != nil {
		return err
	}

	extractLocation := lockitFileInfo.GetExtractLocation()
	if err := extractLocation.ProvideTargetPath(); err != nil {
		return err
	}

	sourceFile := lockitFileInfo.GetGameData().FullFilePath
	targetFile := extractLocation.TargetFile
	
	defer characterTable.Dispose(codeTable)
	
	return Decoder("test", sourceFile, targetFile, codeTable)
} */
