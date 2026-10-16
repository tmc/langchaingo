package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/tools"
)

func main() {
	llm, err := ollama.New(ollama.WithModel("mistral"))
	if err != nil {
		log.Fatal(err)
	}
	memory := memory.NewConversationBuffer()
	ctx := context.TODO()

	streamHandler := callbacks.NewFinalStreamHandler()

	agentTools := []tools.Tool{
		tools.Calculator{},
	}

	executor, err := agents.Initialize(
		llm,
		agentTools,
		agents.ConversationalReactDescription,
		agents.WithMaxIterations(10),
		agents.WithCallbacksHandler(streamHandler),
		agents.WithMemory(memory),
	)
	if err != nil {
		log.Fatal(err)
	}

	// callback function to print streaming result
	callbackFn := func(ctx context.Context, chunk []byte) {
		fmt.Printf("%s", string(chunk))
	}

	// multi-turn conversation, Each turn use independent ReadFromEgress to read the streaming result
	wg := streamHandler.ReadFromEgress(context.TODO(), callbackFn)
	_, err = chains.Run(ctx, executor, "hello, my name is Steve Madden")
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()
	fmt.Printf("\n")

	wg = streamHandler.ReadFromEgress(context.TODO(), callbackFn)
	_, err = chains.Run(ctx, executor, "what is my name ?")
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()
}
