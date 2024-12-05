package tags

type IProcessCodePage interface {
	generateCode(key byte, value string) string
	getMap() map[byte]string
}

type ffxTagsBase struct {}

func (t *ffxTagsBase) processCodePage(processCode IProcessCodePage) []string {
	codePage := make([]string, 0, len(processCode.getMap()))

	for key, value := range processCode.getMap() {
		codePage = append(codePage, processCode.generateCode(key, value))
	}

	return codePage
}