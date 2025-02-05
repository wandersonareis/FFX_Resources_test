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
	ILockitFileOptions interface {
		GetLockitFileOptions() LockitFileOptions
	}

	LockitFileOptions struct {
		NameBase        string
		LineBreaksCount int
		PartsLength     int
		PartsSizes      []int
	}
)

func NewDcpFileOptions(gameVersion int) DcpFileOptions {
	defaultValue := DcpFileOptions{
		NameBase:    "macrodic",
		PartsLength: 5,
	}

	switch gameVersion {
	case 1:
		return defaultValue
	case 2:
		return DcpFileOptions{
			NameBase:    "macrodic",
			PartsLength: 7,
		}
	default:
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
