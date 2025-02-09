package main

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

//go:embed image.png
var image []byte

func main() {
	llm, err := anthropic.New(
		anthropic.WithModel("claude-3-5-sonnet-20240620"),
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
					llms.BinaryPart("image/png", []byte(base64.StdEncoding.EncodeToString(image))),
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
		choices[0].GenerationInfo["InputTokens"],
		choices[0].GenerationInfo["OutputTokens"],
	)
	fmt.Println(choices[0].Content)
	// Output:
	// The string on the box in the image is "LGTM".
}
