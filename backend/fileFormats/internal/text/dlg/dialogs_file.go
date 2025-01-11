package dlg

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"sync"
)

type DialogsFile struct {
	source         interfaces.ISource
	destination    locations.IDestination
	extractorOnce  sync.Once
	extractor      *DlgExtractor
	compressorOnce sync.Once
	compressor     *DlgCompressor
	log            logger.LogHandler
}

func NewDialogs(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	return &DialogsFile{
		source:      source,
		destination: destination,

		log: logger.LogHandler{
			Logger: logger.Get().With().Str("module", "dialogs_file").Logger(),
		},
	}
}

func (d *DialogsFile) Source() interfaces.ISource {
	return d.source
}

func (d *DialogsFile) Extract() error {
	d.extractorOnce.Do(func() {
		d.extractor = NewDlgExtractor(d.source, d.destination)
	})

	if err := d.extractor.Extract(); err != nil {
		return err
	}

	d.log.LogInfo("Dialog file extracted: %s", d.source.Get().Name)

	return nil
}

func (d *DialogsFile) Compress() error {
	d.compressorOnce.Do(func() {
		d.compressor = NewDlgCompressor(d.source, d.destination)
	})

	if err := d.compressor.Compress(); err != nil {
		return err
	}

	d.log.LogInfo("Dialog file compressed: %s", d.destination.Import().Get().GetTargetFile())

	return nil
}
