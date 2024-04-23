package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
)

//go:embed image.png
var image []byte

func main() {
	// As a prerequisite, you need to add model access permissions for the Anthropic Claude3 Haiku model in the AWS Region where you are running.
	// For more information, see https://docs.aws.amazon.com/bedrock/latest/userguide/model-access.html.
	// Specify the AWS Region and Credentials in the standard AWS SDK way.
	llm, err := bedrock.New(
		bedrock.WithModel(bedrock.ModelAnthropicClaudeV3Haiku),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	resp, err := llm.GenerateContent(
		ctx,
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					// For images, you can use image formats such as image/png, image/jpeg, image/gif, image/webp.
					// Please change according to the actual byte array to be given.
					// for more detailes, see this https://docs.anthropic.com/claude/reference/messages_post
					llms.BinaryPart("image/png", image),
					llms.TextPart("Please tell me the string on the box."),
				},
			},
		},
		llms.WithMaxTokens(1000),
		llms.WithTemperature(0.1),
		llms.WithTopP(1.0),
		llms.WithTopK(100),
	)
	if err != nil {
		log.Fatal(err)
	}
	choices := resp.Choices
	if len(choices) < 1 {
		log.Fatal("empty response from model")
	}
	log.Printf(
		"input_tokens: %d, output_tokens: %d",
		choices[0].GenerationInfo["input_tokens"],
		choices[0].GenerationInfo["output_tokens"],
	)
	fmt.Println(choices[0].Content)
	//Output:
	//The string on the box in the image is "LGTM".
}
