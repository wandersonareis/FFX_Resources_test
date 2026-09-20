package objectsfile

import (
	"fmt"
	"sync"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

var (
	populateOnce sync.Once
	populateDone bool
)

type PopulateOptions struct {
	Force  bool
	OnSkip func(key string, err error)
}

type PopulateOption func(*PopulateOptions)

func WithForce(force bool) PopulateOption {
	return func(o *PopulateOptions) { o.Force = force }
}

func WithOnSkip(fn func(key string, err error)) PopulateOption {
	return func(o *PopulateOptions) { o.OnSkip = fn }
}

func InitAndPopulateObjectFileStore(opts ...PopulateOption) (loaded, skipped int) {
	var o PopulateOptions
	for _, fn := range opts {
		fn(&o)
	}
	if !o.Force && populateDone {
		return 0, 0
	}
	done := false
	populateOnce.Do(func() {
		loaded, skipped = populateObjectFileStore(&o)
		populateDone = true
		done = true
	})
	if done {
		return loaded, skipped
	}
	if o.Force {
		loaded, skipped = populateObjectFileStore(&o)
		populateDone = true
	}
	return loaded, skipped
}

func RefreshObjectFileStore(opts ...PopulateOption) (loaded, skipped int) {
	opts = append(opts, WithForce(true))
	return InitAndPopulateObjectFileStore(opts...)
}

func populateObjectFileStore(o *PopulateOptions) (loaded, skipped int) {
	for key, layout := range FileLayouts {
		f, err := LoadObjectFile(layout)
		if err != nil {
			skipped++
			common.LogVerbose("[ObjectStore] pular %s: %v", key, err)
			if o.OnSkip != nil {
				o.OnSkip(key, err)
			}
			continue
		}
		if f == nil || f.GetObjects() == nil || f.GetObjects().IsEmpty() {
			skipped++
			common.LogVerbose("[ObjectStore] pular %s: no objects loaded or empty", key)
			if o.OnSkip != nil {
				o.OnSkip(key, nil)
			}
			continue
		}
		ObjectFileDataStore.Register(key, f)
		loaded++
	}
	common.LogInfo("[ObjectStore] populado: %d carregados, %d pulados", loaded, skipped)
	return loaded, skipped
}

func ResetObjectFileStoreForTest() {
	ObjectFileDataStore.Clear()
	populateOnce = sync.Once{}
	populateDone = false
}

func LoadObjectFile(l FileLayout) (datastore.IBinaryFile, error) {
	patternPath := l.PatternPath()
	var creator CreatorFunc
	if l.Creator != nil {
		creator = l.Creator
	} else {
		layouts := LayoutSet{l.Version: l.Fields}
		creator = func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
			return newKeyedStringStore(cBytes, sBytes, hLen, lang, l.Version, layouts, l.FileName, l.Formatter)
		}
	}
	binaryDataFile := NewObjectBinaryFileStore(
		patternPath,
		creator,
		common.DefaultLocalization,
		l.Version,
	)
	if err := binaryDataFile.LoadFromBinary(); err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", patternPath, err)
	}
	if objects := binaryDataFile.GetObjects(); objects != nil && !objects.IsEmpty() {
		datastore.Commands = objects
	}
	return binaryDataFile, nil
}

func LoadObjectFileFromStore(version common.GameVersion, patternPath string) (datastore.IBinaryFile, error) {
	key := FileLayoutKey(version, patternPath)
	layout, ok := FileLayoutFor(version, patternPath)
	if !ok {
		return nil, fmt.Errorf("no layout registered for key %s", key)
	}
	return LoadObjectFileFromStoreByLayout(layout)
}

// LoadObjectFileFromStoreByLayout carrega o arquivo do layout e registra a
// instância no ObjectFileDataStore. Recebe o layout (versão + dir + arquivo)
// em vez das primitivas separadas.
func LoadObjectFileFromStoreByLayout(l FileLayout) (datastore.IBinaryFile, error) {
	key := FileLayoutKey(l.Version, l.PatternPath())
	binFile, err := LoadObjectFile(l)
	if err != nil {
		return nil, err
	}
	ObjectFileDataStore.Register(key, binFile)
	return binFile, nil
}
