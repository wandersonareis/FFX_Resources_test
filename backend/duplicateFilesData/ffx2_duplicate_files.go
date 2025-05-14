package duplicateFilesData

type Ffx2DuplicateFiles struct {
	*DuplicateFiles

	ffx2DuplicateFileMap *Ffx2DuplicateFileMap
}

var ffx2DuplicateInstance *Ffx2DuplicateFiles

func NewFfx2DuplicateFiles() *Ffx2DuplicateFiles {
	if ffx2DuplicateInstance == nil {
		ffx2DuplicateInstance = &Ffx2DuplicateFiles{
			DuplicateFiles:       NewDuplicateFiles(),
			ffx2DuplicateFileMap: new(Ffx2DuplicateFileMap),
		}
	}

	return ffx2DuplicateInstance
}

func (f *Ffx2DuplicateFiles) TryFind(key string) []string {
	return f.Find(key)
}

func (df *Ffx2DuplicateFiles) PopulateDuplicatesFiles() {
	df.Add(df.ffx2DuplicateFileMap.GetFfx2EventBika07_235())
	df.Add(df.ffx2DuplicateFileMap.GetFfx2EventBika0700())
	df.Add(df.ffx2DuplicateFileMap.GetFfx2EventBsil0300())
	df.Add(df.ffx2DuplicateFileMap.GetFfx2EventDnfr0100())
	df.Add(df.ffx2DuplicateFileMap.GetFfx2EventDnfr8000())
	df.Add(df.ffx2DuplicateFileMap.GetFfx2EventHiku2800())
	df.Add(df.ffx2DuplicateFileMap.GetFfx2EventHiku2903())
	df.Add(df.ffx2DuplicateFileMap.GetFfx2EventHiku3000())
	df.Add(df.ffx2DuplicateFileMap.GetFfx2EventKlyt0500())
	df.Add(df.ffx2DuplicateFileMap.GetFfx2EventKlyt0900())
	df.Add(df.ffx2DuplicateFileMap.GetFfx2EventStbv1200())
}
