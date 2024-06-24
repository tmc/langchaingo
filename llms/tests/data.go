package tests

import (
	"github.com/cockroachdb/errors"
	"github.com/tmc/langchaingo/llms"
)

const (
	PrimaryEnergyHydroPng = "https://d2908q01vomqb2.cloudfront.net/da4b9237bacccdf19c0760cab7aec4a8359010b0/2024/06/19/primary-energy-hydro.png"
	NewYorkPng            = "https://upload.wikimedia.org/wikipedia/commons/thumb/4/47/Statue_of_Liberty_New_York_2021_%28cropped%29.jpg/202px-Statue_of_Liberty_New_York_2021_%28cropped%29.jpg"
)

type testArgs struct {
	messages      []llms.MessageContent
	tools         []llms.Tool
	toolChoice    llms.ToolChoice
	exceptedTools []llms.Tool
}

func (t *testArgs) Validate() error {
	if len(t.messages) == 0 {
		return errors.New("no messages provided")
	}
	if len(t.tools) > 0 && len(t.exceptedTools) == 0 {
		return errors.New("expected tools not provided")
	}
	return nil
}

var (
	converseWithSystem = testArgs{
		messages: []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					llms.TextPart("You know all about AI."),
				},
			},
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart("Explain AI in 10 words or less."),
				},
			},
		},
	}

	converseWithImage = testArgs{
		messages: []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.ImageURLPart(PrimaryEnergyHydroPng),
					llms.TextPart(
						"Which countries consume more than 1000 TWh from hydropower? Think step by step and look at all regions. Output in JSON.",
					),
				},
			},
		},
	}

	converseWithTools = testArgs{
		messages: []llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextPart("What's the weather in Boston?")},
			},
		},
	}

	converseImageWithTools = testArgs{
		messages: []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.ImageURLPart(NewYorkPng),
					llms.TextPart("Tell me the weather of the location in the image, use Celsius as the unit."),
				},
			},
		},
		tools: availableTools,
	}
)
