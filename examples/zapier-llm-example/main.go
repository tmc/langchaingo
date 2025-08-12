package main

import (
	"context"
	"fmt"

	"github.com/vendasta/langchaingo/agents"
	"github.com/vendasta/langchaingo/chains"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/tools"
	"github.com/vendasta/langchaingo/tools/zapier"
)

func main() {
	ctx := context.Background()

	llm, err := openai.New()
	if err != nil {
		panic(err)
	}

	// set env variable ZAPIER_NLA_API_KEY to your Zapier API key

	// get all the available zapier NLA Tools
	tks, err := zapier.Toolkit(ctx, zapier.ToolkitOpts{
		// APIKey: "SOME_KEY_HERE", Or pass in a key here
		// AccessToken: "ACCESS_TOKEN", this is if your using OAuth
	})
	if err != nil {
		panic(err)
	}

	agentTools := []tools.Tool{
		// define tools here
	}
	// add the zapier tools to the existing agentTools
	agentTools = append(agentTools, tks...)

	// Initialize the agent
	agent := agents.NewOneShotAgent(llm,
		agentTools,
		agents.WithMaxIterations(3))
	executor := agents.NewExecutor(agent)

	// run a chain with the executor and defined input
	input := "Get the last email from noreply@github.com"
	answer, err := chains.Run(context.Background(), executor, input)
	if err != nil {
		panic(err)
	}
	fmt.Println(answer)
}
