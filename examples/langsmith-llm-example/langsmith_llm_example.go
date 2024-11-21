package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/langsmith"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

var flagLangchainApiKey = flag.String("langchain-api-key", "", "sets langchain API Key, if value not set will check LANGCHAIN_API_KEY")
var flagLangchainProject = flag.String("langchain-project", "langchain_go_example", "sets langchain  project, if value not set will check LANGCHAIN_PROJECT")

func main() {
	ctx := context.Background()

	llm, err := openai.New(openai.WithModel("gpt-4o"))
	if err != nil {
		log.Fatal(err)
	}

	langchainAPIKey := getFlagOrEnv(flagLangchainApiKey, "LANGCHAIN_API_KEY")
	if err != nil {
		log.Fatal("langchain API Key not set")
	}

	langchainProject := getFlagOrEnv(flagLangchainProject, "LANGCHAIN_PROJECT")
	if err != nil {
		log.Fatal("langchain Project not set")
	}

	fmt.Println("Using langchain project: ", langchainProject)

	logger := &logger{}
	langsmithClient, err := langsmith.NewClient(
		langsmith.WithAPIKey(langchainAPIKey),
		langsmith.WithClientLogger(logger),
		langsmith.WithAPIURL("https://api.smith.langchain.com"),
	)
	if err != nil {
		log.Fatal(err)
	}

	langsmithTracer, err := langsmith.NewTracer(
		langsmith.WithLogger(logger),
		langsmith.WithProjectName(langchainProject),
		langsmith.WithClient(langsmithClient),
		langsmith.WithRunId(uuid.NewString()),
	)
	if err != nil {
		log.Fatal(err)
	}

	translatePrompt := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate("You are a translation expert", nil),
		prompts.NewHumanMessagePromptTemplate("Translate the following text from {{.inputLanguage}} to {{.outputLanguage}}. {{.text}}", nil),
	})

	// Initiate your llm chain with the langsmith tracer
	llmChain := chains.NewLLMChain(llm, translatePrompt, chains.WithCallback(langsmithTracer))

	// To get full tracing we must pass through the Static Call function
	outputValues, err := chains.Call(ctx, llmChain, map[string]any{
		"inputLanguage":  "English",
		"outputLanguage": "French",
		"text":           "I love programming.",
	})
	if err != nil {
		log.Fatal(err)
	}

	cnt, err := json.Marshal(outputValues)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(cnt))
}

func getFlagOrEnv(flagValue *string, envName string) string {
	if flagValue == nil {
		return os.Getenv(envName)
	}
	if *flagValue == "" {
		return os.Getenv(envName)
	}
	return *flagValue
}

type logger struct {
}

func (l *logger) Debugf(format string, v ...interface{}) {
	fmt.Printf("[DEBUG] "+format, v...)
}

func (l *logger) Errorf(format string, v ...interface{}) {
	fmt.Printf("[ERROR] "+format, v...)

}

func (l *logger) Infof(format string, v ...interface{}) {
	fmt.Printf("[INFO] "+format, v...)

}

func (l *logger) Warnf(format string, v ...interface{}) {
	fmt.Printf("[WARN] "+format, v...)
}
