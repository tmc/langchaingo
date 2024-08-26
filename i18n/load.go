package i18n

import (
	"fmt"
	"io/fs"
	"log"
)

func mustLoad(lang Lang, kindFolder, filename string) string {
	langFolderMap := map[Lang]string{
		EN: "EN",
		ZH: "ZH",
	}
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
