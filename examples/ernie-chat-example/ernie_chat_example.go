package main

import (
	"context"
	"fmt"
	"log"
	
	"github.com/yincongcyincong/langchaingo/llms"
	"github.com/yincongcyincong/langchaingo/llms/ernie"
)

func main() {
	llm, err := ernie.New(
		ernie.WithModelName(ernie.ModelNameERNIEBot),
		// Fill in your AK and SK here.
		ernie.WithAKSK("ak", "sk"),
		// Use an external cache for the access token.
		ernie.WithAccessToken("accesstoken"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a company branding design wizard."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What would be a good company name a company that makes colorful socks?"),
	}
	completion, err := llm.GenerateContent(ctx, content, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		fmt.Print(string(chunk))
		return nil
	}))
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println(completion)
}
