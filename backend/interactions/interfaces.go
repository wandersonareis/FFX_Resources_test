package interactions

type IFileProcessor interface {
	GetFileInfo() IGameDataInfo
	Extract()
	Compress()
}

type TreeNode struct {
	Key      string        `json:"key"`
	Label    string        `json:"label"`
	Data     GameDataInfo `json:"data"`
	Icon     string        `json:"icon"`
	Children []TreeNode    `json:"children"`
}
