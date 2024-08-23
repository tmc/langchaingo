package i18n

import (
	"encoding/json"
	"log"
	"sync"
)

func AgentsMustLoad(lang Lang, filename string) string {
	return mustLoad(lang, "agents", filename)
}

// nolint:gochecknoglobals
var langAgentsPhraseMap sync.Map

func AgentsMustPhrase(lang Lang, key string) string {
	var agentsPhraseMap map[string]string
	valAny, ok := langAgentsPhraseMap.Load(lang)
	if ok {
		agentsPhraseMap, _ = valAny.(map[string]string)
	} else {
		s := AgentsMustLoad(lang, "phrase.json")
		agentsPhraseMap = make(map[string]string)
		if err := json.Unmarshal([]byte(s), &agentsPhraseMap); err != nil {
			log.Panic("unmarshal phrase failed:", err)
		}
		langAgentsPhraseMap.Store(lang, agentsPhraseMap)
	}
	val, ok := agentsPhraseMap[key]
	if !ok {
		log.Panic("there is no such key in phrase:", key)
	}
	return val
}
