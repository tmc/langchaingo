package i18n

import (
	"fmt"
	"io/fs"
	"log"
)

var langFolderMap = map[Lang]string{
	EN: "EN",
	ZH: "ZH",
}

func mustLoad(lang Lang, kindFolder, filename string) string {
	langFolder, ok := langFolderMap[lang]
	if !ok {
		log.Panic("unknown language: ", lang)
	}
	filepath := fmt.Sprintf("templates/%s/%s/%s", kindFolder, langFolder, filename)
	b, err := fs.ReadFile(agentsTpls, filepath)
	if err != nil {
		log.Panic("read file failed: ", err)
	}
	return string(b)
}
