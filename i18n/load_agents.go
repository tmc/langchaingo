package i18n

import (
	"embed"
	"encoding/json"
	"log"
)

//go:embed templates/agents/*
var agentsTpls embed.FS

func AgentsMustLoad(lang Lang, filename string) string {
	return mustLoad(lang, "agents", filename)
}

var (
	// nolint:gochecknoglobals
	agentsPhraseMap map[string]string
	// nolint:gochecknoglobals
	agentsPhraseInit bool
)

func AgentsMustPhrase(lang Lang, key string) string {
	if !agentsPhraseInit {
		s := AgentsMustLoad(lang, "phrase.json")
		agentsPhraseMap = make(map[string]string)
		if err := json.Unmarshal([]byte(s), &agentsPhraseMap); err != nil {
			log.Panic("unmarshal phrase failed:", err)
		}
		agentsPhraseInit = true
	}
	val, ok := agentsPhraseMap[key]
	if !ok {
		log.Panic("there is no such key in phrase:", key)
	}
	return val
}
