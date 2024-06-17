package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/serpapi"
	"os"
)

func main() {
	ctx := context.Background()

	// Making sure API keys exist
	serpapiKey := os.Getenv("SERPAPI_API_KEY")
	if serpapiKey == "" {
		panic("SERPAPI_API_KEY is required")
	}

	openAIApiKey := os.Getenv("OPENAI_API_KEY")
	if openAIApiKey == "" {
		panic("OPENAI_API_KEY is required")
	}

	searchTool, err := serpapi.New(serpapi.WithAPIKey(serpapiKey))
	if err != nil {
		panic(err)
	}

	// Constructing a list of tools LLM can reference
	modelTools := []tools.Tool{
		searchTool,
	}

	toolNamesMap := make(map[string]tools.Tool)
	for _, tool := range modelTools {
		toolNamesMap[tool.Name()] = tool
	}

	// Specifying OpenAI LLM to be Openai GPT-3.5 Turbo
	llm, err := openai.New(
		openai.WithModel("gpt-3.5-turbo"),
		openai.WithToken(openAIApiKey),
	)
	if err != nil {
		panic(err)
	}

	// Conversation history for the LLM
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "Which country has greater population in 2024 Germany or France?"),
	}

	res, err := llm.GenerateContent(ctx, messages, llms.WithTools(prepareTools(modelTools)))
	if err != nil {
		panic(err)
	}

	// Add model response to the conversation history including tool call requests
	assistantResponse := llms.TextParts(llms.ChatMessageTypeAI, res.Choices[0].Content)
	for _, tc := range res.Choices[0].ToolCalls {
		assistantResponse.Parts = append(assistantResponse.Parts, tc)
	}
	messages = append(messages, assistantResponse)

	// Execute tools and collect responses
	for _, toolCall := range res.Choices[0].ToolCalls {
		tool, ok := toolNamesMap[toolCall.FunctionCall.Name]
		if !ok {
			panic("could not find tool in dictionary")
		}

		var toolArgs any
		err = json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &toolArgs)
		if err != nil {
			panic(err)
		}

		// execute the tool calls
		toolRes, err := tool.Call(ctx, toolArgs)
		if err != nil {
			panic(err)
		}

		toolResponse := llms.ToolCallResponse{
			ToolCallID: toolCall.ID,
			Name:       toolCall.FunctionCall.Name,
			Content:    toolRes,
		}

		message := llms.MessageContent{
			Role:  llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{toolResponse},
		}

		messages = append(messages, message)
	}

	// Call model again with tool call results
	res, err = llm.GenerateContent(ctx, messages, llms.WithTools(prepareTools(modelTools)))
	if err != nil {
		panic(err)
	}

	fmt.Println(res.Choices[0].Content)
}

func prepareTools(tools []tools.Tool) []llms.Tool {
	llmTools := make([]llms.Tool, len(tools))

	for i, tool := range tools {
		llmTools[i] = llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Schema(),
			},
		}
	}

	return llmTools
}
