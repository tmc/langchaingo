package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
)

//go:embed templates/*
var tpls embed.FS

func mustLoad(lang Lang, kindFolder, filename string) string {
	langFolderMap := map[Lang]string{
		EN: "en",
		ZH: "zh",
	}
	langFolder, ok := langFolderMap[lang]
	if !ok {
		log.Panic("unknown language: ", lang)
	}
	filepath := fmt.Sprintf("templates/%s/%s/%s", kindFolder, langFolder, filename)
	b, err := fs.ReadFile(tpls, filepath)
	if err != nil {
		log.Panic("read file failed: ", err)
	}
	return string(b)
}
