package core

type (
	IDcpFileOptions interface {
		GetDcpFileOptions() DcpFileOptions
	}

	DcpFileOptions struct {
		NameBase    string
		PartsLength int
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
		return defaultValue
	}
}

func NewLockitFileOptions(gameVersion int) LockitFileOptions {
	switch gameVersion {
	case 2:
		sizes := []int{80, 88, 90, 93, 94, 95, 102, 1223, 1224, 1230, 1232, 1233, 1240, 1241, 1502, 1534}
		return LockitFileOptions{
			NameBase:        "loc_kit_ps3",
			LineBreaksCount: 1696,
			PartsLength:     len(sizes) + 1,
			PartsSizes:      sizes,
		}
	default:
		// FFX não usa Lockit
		return LockitFileOptions{}
	}
}
