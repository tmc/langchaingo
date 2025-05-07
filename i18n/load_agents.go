package i18n

import (
	"encoding/json"
	"log"
	"sync"
)

// AgentsMustLoad loads file of agents by the given language and filename.
// Will panic if any problem occurs, including unsupported language,
// unknown filename and other problems.
func AgentsMustLoad(lang Lang, filename string) string {
	return mustLoad(lang, "agents", filename)
}

// nolint:gochecknoglobals
var langAgentsPhraseMap sync.Map

// AgentsMustPhrase loads phrase of agents by the given language and key.
// Will panic if any problem occurs, including unsupported language,
// unknown key and other problems.
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
