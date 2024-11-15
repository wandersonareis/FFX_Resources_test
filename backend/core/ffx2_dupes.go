package core

import "ffxresources/backend/data"

type Ffx2Duplicate struct {
	Duplicate
}

var ffx2DuplicateInstance *Ffx2Duplicate

func NewFfx2Duplicate() *Ffx2Duplicate {
	if ffx2DuplicateInstance == nil {
		ffx2DuplicateInstance = &Ffx2Duplicate{
			Duplicate: *NewDuplicate(),
		}
	}

	return ffx2DuplicateInstance
}

func (f *Ffx2Duplicate) TryFind(key string) []string {
	return f.Find(key)
}

func (f *Ffx2Duplicate) AddFfx2TextDuplicate() {
	f.AddFromData(data.Ffx2_btl_bika07_235)
	f.AddFromData(data.Ffx2_event_bika0700)
	f.AddFromData(data.Ffx2_event_bsil0300)
	f.AddFromData(data.Ffx2_event_dnfr0100)
	f.AddFromData(data.Ffx2_event_dnfr8000)
	f.AddFromData(data.Ffx2_event_hiku2800)
	f.AddFromData(data.Ffx2_event_hiku2903)
	f.AddFromData(data.Ffx2_event_hiku3000)
	f.AddFromData(data.Ffx2_event_klyt0500)
	f.AddFromData(data.Ffx2_event_klyt0900)
	f.AddFromData(data.Ffx2_event_stbv1200)
}
