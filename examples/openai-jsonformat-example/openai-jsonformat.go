package main

import (
	"context"
	"github.com/yincongcyincong/langchaingo/llms"
	"github.com/yincongcyincong/langchaingo/llms/openai"
	"log"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	fomat := &openai.ResponseFormat{
		Type: "json_schema",
		JSONSchema: &openai.ResponseFormatJSONSchema{
			Name: "object",
			Schema: &openai.ResponseFormatJSONSchemaProperty{
				Type: "object",
				Properties: map[string]*openai.ResponseFormatJSONSchemaProperty{
					"name": {
						Type:        "string",
						Description: "The name of the user",
					},
					"age": {
						Type:        "integer",
						Description: "The age of the user",
					},
					"role": {
						Type:        "string",
						Description: "The role of the user",
					},
				},
				AdditionalProperties: false,
				Required:             []string{"name", "age", "role"},
			},
			Strict: true,
		},
	}
	llm, err := openai.New(openai.WithModel("gpt-4o"), openai.WithResponseFormat(fomat))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are an expert at structured data extraction. You will be given unstructured text from a research paper and should convert it into the given structure."),
		llms.TextParts(llms.ChatMessageTypeHuman, "please tell me the most famous people in history"),
	}
	
	completion, err := llm.GenerateContent(ctx, content, llms.WithJSONMode())
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(completion.Choices[0].Content)
}
