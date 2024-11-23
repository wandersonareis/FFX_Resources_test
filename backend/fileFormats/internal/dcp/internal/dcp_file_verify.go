package internal

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
	"path/filepath"
)

type DcpFileVerify struct {
	*base.FormatsBase
}

func NewDcpFileVerify(dataInfo interactions.IGameDataInfo) *DcpFileVerify {
	return &DcpFileVerify{
		FormatsBase: base.NewFormatsBase(dataInfo),
	}
}

func (lv *DcpFileVerify) VerifyExtract(extractLocation *interactions.ExtractLocation, options *interactions.DcpFileOptions) error {
	errChan := make(chan error, 10)

	lv.Log.Info().Msgf("Verifying splited macrodic file: %s", extractLocation.TargetPath)

	if err := lv.verifyDcpFileParts(extractLocation.TargetPath, options, errChan); err != nil {
		errChan <- fmt.Errorf("error when verifying splited macrodic file: %w", err)
		return <-errChan
	}

	if err := lv.verifyDcpTextParts(extractLocation.TargetPath, options, errChan); err != nil {
		errChan <- fmt.Errorf("error when verifying splited macrodic text parts: %w", err)
		return <-errChan
	}

	for {
		select {
		case err := <-errChan:
			if err != nil {
				lv.Log.Error().Err(err).Msg("error when verifying splited macrodic file")
				extractLocation.DisposeTargetPath()
				return err
			}
		default:
			if len(errChan) == 0 {
				lv.Log.Info().Msgf("Macrodic file splited successfully to: %s", extractLocation.TargetFileName)

				return nil
			}
		}
	}
}

func (lv *DcpFileVerify) VerifyCompress(dataInfo interactions.IGameDataInfo, options *interactions.DcpFileOptions) error {
	errChan := make(chan error, 10)

	lv.Log.Info().Msgf("Verifying reimported macrodic file: %s", dataInfo.GetImportLocation().TargetFile)

	if err := dataInfo.GetImportLocation().Validate(); err != nil {
		errChan <- fmt.Errorf("reimport file not exists: %s | %w", dataInfo.GetImportLocation().TargetFile, err)
		return <-errChan
	}

	if err := lv.verifyDcpTextParts(dataInfo.GetTranslateLocation().TargetPath, options, errChan); err != nil {
		errChan <- fmt.Errorf("error when verifying reimported macrodic text parts: %w", err)
		return <-errChan
	}

	if err := lv.valideDcpFile(dataInfo.GetImportLocation().TargetFile, options, errChan); err != nil {
		errChan <- fmt.Errorf("error when validating reimported macrodic file: %w", err)
		return <-errChan
	}

	for {
		select {
		case err := <-errChan:
			if err != nil {
				lv.Log.Error().Err(err).Msg("error when verifying reimported macrodic file")
				dataInfo.GetImportLocation().DisposeTargetFile()
				return err
			}
		default:
			if len(errChan) == 0 {
				lv.Log.Info().Msgf("Macrodic file reimported successfully to: %s", dataInfo.GetImportLocation().TargetFile)

				return nil
			}
		}
	}
}

func (lv *DcpFileVerify) verifyDcpFileParts(targetPath string, options *interactions.DcpFileOptions, errChan chan error) error {
	parts := []DcpFileParts{}

	if err := util.FindFileParts(&parts, targetPath, util.DCP_FILE_PARTS_PATTERN, NewDcpFileParts); err != nil {
		errChan <- fmt.Errorf("error when finding macrodic parts in %s", targetPath)
		return <-errChan
	}

	if len(parts) != options.PartsLength {
		lv.Log.Error().Int("Current parts", len(parts)).Int("Expected parts", options.PartsLength).Msg("invalid number of parts")
		errChan <- fmt.Errorf("invalid number of parts: Expected parts: %d Won: %d", len(parts), options.PartsLength)
		return <-errChan
	}

	for _, part := range parts {
		if part.GetGameData().Size == 0 {
			errChan <- fmt.Errorf("invalid size for part: %s", part.GetGameData().Name)
			return <-errChan
		}
	}

	return nil
}

func (lv *DcpFileVerify) verifyDcpTextParts(targetPath string, options *interactions.DcpFileOptions, errChan chan error) error {
	parts := &[]DcpFileParts{}

	worker := common.NewWorker[DcpFileParts]()
	defer worker.Close()

	worker.Execute(func() error {
		return util.FindFileParts(parts, targetPath, util.DCP_TXT_PARTS_PATTERN, NewDcpFileParts)
	}, lv.Log, fmt.Sprintf("error when finding macrodic text parts in %s", targetPath), errChan)

	if len(*parts) != options.PartsLength {
		errChan <- fmt.Errorf("text parts: %d Expected text parts: %d", len(*parts), options.PartsLength)
		return <-errChan
	}

	worker.ForEach(*parts, func(i int, part DcpFileParts) error {
		if common.CountSegments(part.GetTranslateLocation().TargetFile) == 0 {
			errChan <- fmt.Errorf("invalid number of segments in text part: %s", part.GetTranslateLocation().TargetFile)
			return nil
		}
		return nil
	})

	lv.Log.Info().Msgf("Macrodic text parts verified successfully in: %s", targetPath)

	return nil
}

func (lv *DcpFileVerify) valideDcpFile(dcpFile string, options *interactions.DcpFileOptions, errChan chan error) error {
	tmpDir := common.NewTempProvider().ProvideTempDir()

	tmpInfo := interactions.NewGameDataInfo(dcpFile)
	tmpInfo.InitializeLocations(formatters.NewTxtFormatter())

	tmpInfo.GetExtractLocation().TargetPath = tmpDir
	defer tmpInfo.GetExtractLocation().DisposeTargetPath()

	worker := common.NewWorker[DcpFileParts]()
	defer worker.Close()

	worker.Execute(func() error {
		return DcpFileXpliter(tmpInfo)
	}, lv.Log, "Error validing macrodic reimported file", errChan)

	parts := &[]DcpFileParts{}

	worker.Execute(func() error {
		return util.FindFileParts(parts, tmpDir, util.DCP_FILE_PARTS_PATTERN, NewDcpFileParts)
	}, lv.Log, fmt.Sprintf("error when finding lockit parts in %s", dcpFile), errChan)

	if len(*parts) != options.PartsLength {
		errChan <- fmt.Errorf("invalid number of parts: Want: %d Expected parts: %d", len(*parts), options.PartsLength)
		return <-errChan
	}

	worker.Execute(func() error {
		return lv.tmpPartsVerify(parts, tmpDir, errChan)
	}, lv.Log, "error when verifying reimported macrodic file", errChan)

	return nil
}

func (dv *DcpFileVerify) tmpPartsVerify(parts *[]DcpFileParts, tmpDir string, errChan chan error) error {
	worker := common.NewWorker[DcpFileParts]()

	worker.ParallelForEach(parts, func(i int, part DcpFileParts) {
		tmpPart := &part
		newPartFile := filepath.Join(tmpDir, part.GetExtractLocation().TargetFileName)

		tmpPart.GetExtractLocation().SetTargetFile(newPartFile)
		tmpPart.GetExtractLocation().SetTargetDirectory(tmpDir)

		if err := dv.comparePartsContent(tmpPart.GetGameData().Name, tmpPart.GetGameData().FullFilePath, tmpPart.GetImportLocation().TargetFile,
			"extracted part is different from imported", errChan); err != nil {
			errChan <- err
			return
		}

		tmpPart.Extract()

		if err := tmpPart.GetExtractLocation().Validate(); err != nil {
			errChan <- fmt.Errorf("error when validating part: %s | %w", part.GetGameData().Name, err)
			return
		}

		if err := tmpPart.GetTranslateLocation().Validate(); err != nil {
			errChan <- fmt.Errorf("error when validating translated part: %s", part.GetGameData().Name)
			return
		}

		if err := dv.comparePartsContent(tmpPart.GetGameData().Name, tmpPart.GetTranslateLocation().TargetFile, tmpPart.GetExtractLocation().TargetFile,
			"translated text is different from reimported", errChan); err != nil {
			errChan <- err
			return
		}
	})

	return nil
}

func (dv *DcpFileVerify) comparePartsContent(fileName, fromFile, toFile, errorMsg string, errChan chan error) error {
	newExtractedPartData, err := os.ReadFile(fromFile)
	if err != nil {
		errChan <- fmt.Errorf("error when reading extracted part: %s", fileName)
		return <-errChan
	}

	importedPartData, err := os.ReadFile(toFile)
	if err != nil {
		errChan <- fmt.Errorf("error when reading imported part: %s", fileName)
		return <-errChan
	}

	if !bytes.Equal(newExtractedPartData, importedPartData) {
		errChan <- fmt.Errorf("%s part: %s Expected size: %d Won size: %d", errorMsg, fileName, len(newExtractedPartData), len(importedPartData))
		return <-errChan
	}

	return nil
}
