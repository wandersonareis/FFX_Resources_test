package models

import "ffxresources/backend/common"

type (
	IDcpFileProperties interface {
		GetNameBase() string
		GetPartsLength() int
	}

	dcpFileOptions struct {
		nameBase    string
		partsLength int
	}

	FFXDcpFile struct {
		dcpFileOptions
	}
)

func NewDcpFileOptions(gameVersion common.GameVersion) IDcpFileProperties {
	switch gameVersion.Normalize() {
	case common.GameVersionFFX2, common.GameVersionLastMiss:
		return NewFFX2DcpFile()
	default:
		return NewFFXDcpFile()
	}
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
