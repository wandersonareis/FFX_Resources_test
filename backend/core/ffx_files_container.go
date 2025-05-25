package core

type (
	lockitFileOptions struct {
		nameBase        string
		lineBreaksCount int
	}

	ILockitFileOptions interface {
		GetNameBase() string
		GetLineBreaksCount() int
		GetPartsLength() int
		GetPartsSizes() []int
	}
)

type FFXLockitFile struct{}

func (ffx *FFXLockitFile) GetNameBase() string {
	return "loc_kit_ps3"
}

func (ffx *FFXLockitFile) GetLineBreaksCount() int {
	return 0 // Ainda não implementado
}

func (ffx *FFXLockitFile) GetPartsLength() int {
	return 0 // Ainda não implementado
}

func (ffx *FFXLockitFile) GetPartsSizes() []int {
	return []int{}
}

type FFX2LockitFile struct {
	lockitFileOptions

	partsSizes [17]int
}

func NewFFX2LockitFile() *FFX2LockitFile {
	return &FFX2LockitFile{
		lockitFileOptions: lockitFileOptions{
			nameBase:        "loc_kit_ps3",
			lineBreaksCount: 1696,
		},
		partsSizes: [17]int{80, 8, 2, 3, 1, 1, 7, 1121, 1, 6, 2, 1, 7, 1, 261, 32, 162},
	}
}

func (ffx2 *FFX2LockitFile) GetNameBase() string {
	return ffx2.nameBase
}

func (ffx2 *FFX2LockitFile) GetLineBreaksCount() int {
	return ffx2.lineBreaksCount
}

func (ffx2 *FFX2LockitFile) GetPartsLength() int {
	return len(ffx2.partsSizes)
}

func (ffx2 *FFX2LockitFile) GetPartsSizes() []int {
	return ffx2.partsSizes[:]
}

func NewLockitFileOptions(gameVersion int) ILockitFileOptions {
	switch gameVersion {
	case 1:
		return &FFXLockitFile{}
	case 2:
		return NewFFX2LockitFile()
	default:
		return &FFXLockitFile{}
	}
}
