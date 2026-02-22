package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vxcontrol/langchaingo/callbacks"
	"github.com/vxcontrol/langchaingo/chains"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
	"github.com/vxcontrol/langchaingo/prompts"
	"github.com/vxcontrol/langchaingo/schema"
	"github.com/vxcontrol/langchaingo/tools"

	"github.com/google/uuid"
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

func (o *OpenAIFunctionsAgent) tools() []llms.Tool {
	res := make([]llms.Tool, 0, len(o.Tools))
	for _, tool := range o.Tools {
		res = append(res, llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters: map[string]any{
					"properties": map[string]any{
						"__arg1": map[string]string{"title": "__arg1", "type": "string"},
					},
					"required": []string{"__arg1"},
					"type":     "object",
				},
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

	var stream streaming.Callback

	if o.CallbacksHandler != nil {
		stream = func(ctx context.Context, chunk streaming.Chunk) error {
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
					Name:       p.Name,
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
	llmOptions := []llms.CallOption{llms.WithTools(o.tools()), llms.WithStreamingFunc(stream)} //nolint:prealloc
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
	messageFormatters := make([]prompts.MessageFormatter, 0, 3+len(opts.extraMessages))
	messageFormatters = append(messageFormatters, prompts.NewSystemMessagePromptTemplate(opts.systemMessage, nil))
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
				Name:    steps[j].Action.Tool,
				Content: steps[j].Observation,
			})
		}
	}

	return messages
}

func (o *OpenAIFunctionsAgent) ParseOutput(contentResp *llms.ContentResponse) (
	[]schema.AgentAction, *schema.AgentFinish, error,
) {
	if contentResp == nil {
		return nil, nil, fmt.Errorf("content response is nil")
	}

	if len(contentResp.Choices) == 0 {
		return nil, nil, fmt.Errorf("content response has no choices")
	}

	var actions []schema.AgentAction

	for _, choice := range contentResp.Choices {
		content := choice.Content

		if len(choice.ToolCalls) == 0 && choice.FuncCall == nil {
			return nil, &schema.AgentFinish{
				ReturnValues: map[string]any{
					"output": content,
				},
				Log: content,
			}, nil
		}

		if len(choice.ToolCalls) == 0 && choice.FuncCall != nil {
			functionCall := choice.FuncCall
			functionName := functionCall.Name
			toolInputStr := functionCall.Arguments
			action, err := o.parseToolCalls("", functionName, toolInputStr, content)
			if err != nil {
				return nil, nil, err
			}
			actions = append(actions, action)
			continue
		}

		for _, toolCall := range choice.ToolCalls {
			if toolCall.FunctionCall == nil {
				continue
			}

			functionName := toolCall.FunctionCall.Name
			toolInputStr := toolCall.FunctionCall.Arguments
			action, err := o.parseToolCalls(toolCall.ID, functionName, toolInputStr, content)
			if err != nil {
				return nil, nil, err
			}
			actions = append(actions, action)
		}
	}

	return actions, nil, nil
}

func (o *OpenAIFunctionsAgent) parseToolCalls(id, name, args, content string) (schema.AgentAction, error) { //nolint:unparam
	argsMap := make(map[string]any, 0)
	err := json.Unmarshal([]byte(args), &argsMap)
	if err != nil {
		// If it's not valid JSON, it might be a raw expression for the calculator
		// Try to use it directly as tool input
		return schema.AgentAction{ //nolint:nilerr
			Tool:      name,
			ToolInput: args,
			Log:       fmt.Sprintf("Invoking: %s with %s\n", name, args),
			ToolID:    id, // Legacy function calls don't have tool IDs
		}, nil
	}

	// extract the first argument from the tool call if it exists
	if arg1, ok := argsMap["__arg1"]; ok {
		argCheck, ok := arg1.(string)
		if ok {
			args = argCheck
		}
	}

	if id == "" {
		id = strings.ReplaceAll(uuid.New().String(), "-", "")
	}
	if content != "" {
		content = fmt.Sprintf("responded: %s\n", content)
	} else {
		content = "\n"
	}

	return schema.AgentAction{
		Tool:      name,
		ToolInput: args,
		Log:       fmt.Sprintf("Invoking: %s with %s\n%s\n", name, args, content),
		ToolID:    id,
	}, nil
}
