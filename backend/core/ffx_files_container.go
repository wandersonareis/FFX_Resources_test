package core

type (
	IDcpFileOptions interface {
		GetNameBase() string
		GetPartsLength() int
	}

	dcpFileOptions struct {
		nameBase    string
		partsLength int
	}

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



type FFXDcpFile struct {
	dcpFileOptions
}

func NewFFXDcpFile() *FFXDcpFile {
	return &FFXDcpFile{
		dcpFileOptions: dcpFileOptions{
			nameBase:    "macrodic",
			partsLength: 5,
		},
	}
}

func (ffxDcp *FFXDcpFile) GetNameBase() string {
	return ffxDcp.nameBase
}

func (ffxDcp *FFXDcpFile) GetPartsLength() int {
	return ffxDcp.partsLength
}

type FFX2DcpFile struct {
	dcpFileOptions
}

func NewFFX2DcpFile() *FFX2DcpFile {
	return &FFX2DcpFile{
		dcpFileOptions: dcpFileOptions{
			nameBase:    "macrodic",
			partsLength: 7,
		},
	}
}

func (ffx2Dcp *FFX2DcpFile) GetNameBase() string {
	return ffx2Dcp.nameBase
}

func (ffx2Dcp *FFX2DcpFile) GetPartsLength() int {
	return ffx2Dcp.partsLength
}

func NewDcpFileOptions(gameVersion int) IDcpFileOptions {
	switch gameVersion {
	case 1:
		return NewFFXDcpFile()
	case 2:
		return NewFFX2DcpFile()
	default:
		return NewFFXDcpFile()
	}
}
