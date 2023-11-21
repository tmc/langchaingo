package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

var (
	flagImagePath = flag.String("image", "-", "path to image to send to model")
)

func main() {
	flag.Parse()
	llm, err := openai.NewChat(openai.WithModel("gpt-4-vision-preview"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	base64Image, err := loadImageBase64(*flagImagePath)
	if err != nil {
		log.Fatal(err)
	}
	completion, err := llm.Call(ctx, []schema.ChatMessage{
		schema.CompoundChatMessage{
			Type: schema.ChatMessageTypeHuman,
			Parts: []schema.ChatMessageContentPart{
				schema.ChatMessageContentPartText{
					Type: "text",
					Text: "What is in this image?",
				},
				schema.ChatMessageContentPartImage{
					Type: "image_url",
					ImageURL: schema.ChatMessageContentPartImageURL{
						URL: base64Image,
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion)

}

func loadImageBase64(path string) (string, error) {
	f, err := pathToReader(path)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	// Determine the content type of the image file
	mimeType := http.DetectContentType(data)

	var base64Encoding string
	// Prepend the appropriate URI scheme header depending
	// on the MIME type
	switch mimeType {
	case "image/jpeg":
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding += "data:image/png;base64,"
	}
	base64Encoding += base64.StdEncoding.EncodeToString(data)
	return base64Encoding, nil
}

func pathToReader(path string) (io.ReadCloser, error) {
	if path == "-" {
		return os.Stdin, nil
	}
	return os.Open(path)
}
