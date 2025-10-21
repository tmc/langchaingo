package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

// agentScratchpad "agent_scratchpad" for the agent to put its thoughts in.
const agentScratchpad = "agent_scratchpad"

// OpenAIFunctionsAgent is an Agent driven by OpenAIs function powered API.
type OpenAIFunctionsAgent struct {
	// LLM is the llm used to call with the values. The llm should have an
	// input called "agent_scratchpad" for the agent to put its thoughts in.
	LLM    llms.Model
	Prompt prompts.FormatPrompter
	// Chain chains.Chain
	// Tools is a list of the tools the agent can use.
	Tools []tools.Tool
	// Output key is the key where the final output is placed.
	OutputKey string
	// CallbacksHandler is the handler for callbacks.
	CallbacksHandler callbacks.Handler
}

var _ Agent = (*OpenAIFunctionsAgent)(nil)

// NewOpenAIFunctionsAgent creates a new OpenAIFunctionsAgent.
func NewOpenAIFunctionsAgent(llm llms.Model, tools []tools.Tool, opts ...Option) *OpenAIFunctionsAgent {
	options := openAIFunctionsDefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &OpenAIFunctionsAgent{
		LLM:              llm,
		Prompt:           createOpenAIFunctionPrompt(options),
		Tools:            tools,
		OutputKey:        options.outputKey,
		CallbacksHandler: options.callbacksHandler,
	}
}

func (o *OpenAIFunctionsAgent) functions() []llms.FunctionDefinition {
	res := make([]llms.FunctionDefinition, 0)
	for _, tool := range o.Tools {
		res = append(res, llms.FunctionDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters: map[string]any{
				"properties": map[string]any{
					"__arg1": map[string]string{"title": "__arg1", "type": "string"},
				},
				"required": []string{"__arg1"},
				"type":     "object",
			},
		})
	}
	return res
}

// Plan decides what action to take or returns the final result of the input.
func (o *OpenAIFunctionsAgent) Plan(
	ctx context.Context,
	intermediateSteps []schema.AgentStep,
	inputs map[string]string,
	options ...chains.ChainCallOption,
) ([]schema.AgentAction, *schema.AgentFinish, error) {
	fullInputs := make(map[string]any, len(inputs))
	for key, value := range inputs {
		fullInputs[key] = value
	}
	fullInputs[agentScratchpad] = o.constructScratchPad(intermediateSteps)

	var stream func(ctx context.Context, chunk []byte) error

	if o.CallbacksHandler != nil {
		stream = func(ctx context.Context, chunk []byte) error {
			o.CallbacksHandler.HandleStreamingFunc(ctx, chunk)
			return nil
		}
	}

	prompt, err := o.Prompt.FormatPrompt(fullInputs)
	if err != nil {
		return nil, nil, err
	}

	mcList := make([]llms.MessageContent, len(prompt.Messages()))
	for i, msg := range prompt.Messages() {
		role := msg.GetType()
		text := msg.GetContent()

		var mc llms.MessageContent

		switch p := msg.(type) {
		case llms.ToolChatMessage:
			mc = llms.MessageContent{
				Role: role,
				Parts: []llms.ContentPart{llms.ToolCallResponse{
					ToolCallID: p.ID,
					Content:    p.Content,
				}},
			}

		case llms.FunctionChatMessage:
			mc = llms.MessageContent{
				Role: role,
				Parts: []llms.ContentPart{llms.ToolCallResponse{
					Name:    p.Name,
					Content: p.Content,
				}},
			}

		case llms.AIChatMessage:
			if len(p.ToolCalls) > 0 {
				toolCallParts := make([]llms.ContentPart, 0, len(p.ToolCalls))
				for _, tc := range p.ToolCalls {
					toolCallParts = append(toolCallParts, llms.ToolCall{
						ID:           tc.ID,
						Type:         tc.Type,
						FunctionCall: tc.FunctionCall,
					})
				}
				mc = llms.MessageContent{
					Role:  role,
					Parts: toolCallParts,
				}
			} else {
				mc = llms.MessageContent{
					Role:  role,
					Parts: []llms.ContentPart{llms.TextContent{Text: text}},
				}
			}
		default:
			mc = llms.MessageContent{
				Role:  role,
				Parts: []llms.ContentPart{llms.TextContent{Text: text}},
			}
		}
		mcList[i] = mc
	}

	// Build LLM call options, including user-provided options
	llmOptions := []llms.CallOption{llms.WithFunctions(o.functions()), llms.WithStreamingFunc(stream)}
	llmOptions = append(llmOptions, chains.GetLLMCallOptions(options...)...)

	result, err := o.LLM.GenerateContent(ctx, mcList, llmOptions...)
	if err != nil {
		return nil, nil, err
	}

	return o.ParseOutput(result)
}

func (o *OpenAIFunctionsAgent) GetInputKeys() []string {
	chainInputs := o.Prompt.GetInputVariables()

	// Remove inputs given in plan.
	agentInput := make([]string, 0, len(chainInputs))
	for _, v := range chainInputs {
		if v == agentScratchpad {
			continue
		}
		agentInput = append(agentInput, v)
	}

	return agentInput
}

func (o *OpenAIFunctionsAgent) GetOutputKeys() []string {
	return []string{o.OutputKey}
}

func (o *OpenAIFunctionsAgent) GetTools() []tools.Tool {
	return o.Tools
}

func createOpenAIFunctionPrompt(opts Options) prompts.ChatPromptTemplate {
	messageFormatters := []prompts.MessageFormatter{prompts.NewSystemMessagePromptTemplate(opts.systemMessage, nil)}
	messageFormatters = append(messageFormatters, opts.extraMessages...)
	messageFormatters = append(messageFormatters, prompts.NewHumanMessagePromptTemplate("{{.input}}", []string{"input"}))
	messageFormatters = append(messageFormatters, prompts.MessagesPlaceholder{
		VariableName: agentScratchpad,
	})

	tmpl := prompts.NewChatPromptTemplate(messageFormatters)
	return tmpl
}

func (o *OpenAIFunctionsAgent) constructScratchPad(steps []schema.AgentStep) []llms.ChatMessage {
	if len(steps) == 0 {
		return nil
	}

	messages := make([]llms.ChatMessage, 0)

	// Group steps by their position to handle multiple tool calls
	// that might be executed in parallel
	var currentToolCalls []llms.ToolCall
	var currentLog string

	for i, step := range steps {
		// Check if this step is part of a group of parallel tool calls
		// by looking at the log content
		if i == 0 || step.Action.Log != steps[i-1].Action.Log {
			// Start a new group
			if len(currentToolCalls) > 0 {
				// Add the previous group as an AI message
				messages = append(messages, llms.AIChatMessage{
					Content:   currentLog,
					ToolCalls: currentToolCalls,
				})
				// Add tool responses for the previous group
				for j := i - len(currentToolCalls); j < i; j++ {
					messages = append(messages, llms.ToolChatMessage{
						ID:      steps[j].Action.ToolID,
						Content: steps[j].Observation,
					})
				}
				currentToolCalls = nil
			}
			currentLog = step.Action.Log
		}

		// Add this tool call to the current group
		currentToolCalls = append(currentToolCalls, llms.ToolCall{
			ID:   step.Action.ToolID,
			Type: "function",
			FunctionCall: &llms.FunctionCall{
				Name:      step.Action.Tool,
				Arguments: step.Action.ToolInput,
			},
		})
	}

	// Don't forget the last group
	if len(currentToolCalls) > 0 {
		messages = append(messages, llms.AIChatMessage{
			Content:   currentLog,
			ToolCalls: currentToolCalls,
		})
		// Add tool responses for the last group
		for j := len(steps) - len(currentToolCalls); j < len(steps); j++ {
			messages = append(messages, llms.ToolChatMessage{
				ID:      steps[j].Action.ToolID,
				Content: steps[j].Observation,
			})
		}
	}

	return messages
}

func (o *OpenAIFunctionsAgent) ParseOutput(contentResp *llms.ContentResponse) (
	[]schema.AgentAction, *schema.AgentFinish, error,
) {
	if contentResp == nil || len(contentResp.Choices) == 0 {
		return nil, nil, fmt.Errorf("no choices in response")
	}
	choice := contentResp.Choices[0]

	// Check for new-style tool calls first
	if len(choice.ToolCalls) > 0 {
		// Handle multiple tool calls properly
		actions := make([]schema.AgentAction, 0, len(choice.ToolCalls))

		for _, toolCall := range choice.ToolCalls {
			functionName := toolCall.FunctionCall.Name
			toolInputStr := toolCall.FunctionCall.Arguments
			toolInputMap := make(map[string]any, 0)
			err := json.Unmarshal([]byte(toolInputStr), &toolInputMap)

			toolInput := toolInputStr
			if err == nil {
				// Successfully parsed JSON, check for __arg1 pattern
				if arg1, ok := toolInputMap["__arg1"]; ok {
					toolInputCheck, ok := arg1.(string)
					if ok {
						toolInput = toolInputCheck
					}
				}
			}
			// If JSON parsing failed, use the raw string as tool input
			// This handles cases like calculator expressions

			contentMsg := "\n"
			if choice.Content != "" {
				contentMsg = fmt.Sprintf("responded: %s\n", choice.Content)
			}

			actions = append(actions, schema.AgentAction{
				Tool:      functionName,
				ToolInput: toolInput,
				Log:       fmt.Sprintf("Invoking: %s with %s %s", functionName, toolInputStr, contentMsg),
				ToolID:    toolCall.ID,
			})
		}

		return actions, nil, nil
	}

	// Check for legacy function call
	if choice.FuncCall != nil {
		functionCall := choice.FuncCall
		functionName := functionCall.Name
		toolInputStr := functionCall.Arguments
		toolInputMap := make(map[string]any, 0)
		err := json.Unmarshal([]byte(toolInputStr), &toolInputMap)
		if err != nil {
			// If it's not valid JSON, it might be a raw expression for the calculator
			// Try to use it directly as tool input
			return []schema.AgentAction{
				{
					Tool:      functionName,
					ToolInput: toolInputStr,
					Log:       fmt.Sprintf("Invoking: %s with %s\n", functionName, toolInputStr),
					ToolID:    "", // Legacy function calls don't have tool IDs
				},
			}, nil, nil
		}

		toolInput := toolInputStr
		if arg1, ok := toolInputMap["__arg1"]; ok {
			toolInputCheck, ok := arg1.(string)
			if ok {
				toolInput = toolInputCheck
			}
		}

		contentMsg := "\n"
		if choice.Content != "" {
			contentMsg = fmt.Sprintf("responded: %s\n", choice.Content)
		}

		return []schema.AgentAction{
			{
				Tool:      functionName,
				ToolInput: toolInput,
				Log:       fmt.Sprintf("Invoking: %s with %s \n %s \n", functionName, toolInputStr, contentMsg),
				ToolID:    "", // Legacy function calls don't have tool IDs
			},
		}, nil, nil
	}

	// No function/tool call - this is a finish
	return nil, &schema.AgentFinish{
		ReturnValues: map[string]any{
			"output": choice.Content,
		},
		Log: choice.Content,
	}, nil
}
