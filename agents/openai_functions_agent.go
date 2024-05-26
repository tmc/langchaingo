package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
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

func (o *OpenAIFunctionsAgent) tools() []llms.Tool {
	res := make([]llms.Tool, 0)
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
	intermediateMessages []llms.ChatMessage,
) ([]schema.AgentAction, *schema.AgentFinish, []llms.ChatMessage, error) {
	fullInputs := make(map[string]any, len(inputs))
	for key, value := range inputs {
		fullInputs[key] = value
	}
	fullInputs[agentScratchpad] = o.constructScratchPad(intermediateMessages, intermediateSteps)

	var stream func(ctx context.Context, chunk []byte) error

	if o.CallbacksHandler != nil {
		stream = func(ctx context.Context, chunk []byte) error {
			o.CallbacksHandler.HandleStreamingFunc(ctx, chunk)
			return nil
		}
	}

	prompt, err := o.Prompt.FormatPrompt(fullInputs)
	if err != nil {
		return nil, nil, nil, err
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

		case llms.AIChatMessage:
			mc = llms.MessageContent{
				Role: role,
			}
			var contentParts []llms.ContentPart
			for _, toolCall := range p.ToolCalls {
				contentParts = append(contentParts, llms.ToolCall{
					ID:           toolCall.ID,
					Type:         toolCall.Type,
					FunctionCall: toolCall.FunctionCall,
				})
			}
			mc.Parts = contentParts

		default:
			mc = llms.MessageContent{
				Role:  role,
				Parts: []llms.ContentPart{llms.TextContent{Text: text}},
			}
		}
		mcList[i] = mc
	}

	result, err := o.LLM.GenerateContent(ctx, mcList,
		llms.WithTools(o.tools()), llms.WithStreamingFunc(stream))
	if err != nil {
		return nil, nil, nil, err
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

func (o *OpenAIFunctionsAgent) constructScratchPad(intermediateMessages []llms.ChatMessage, steps []schema.AgentStep) []llms.ChatMessage {
	if len(steps) == 0 {
		return nil
	}

	messages := make([]llms.ChatMessage, 0)

	var toolCalls []llms.ToolCall
	for _, message := range intermediateMessages {
		toolIDSet := make(map[string]bool)
		for _, toolCall := range message.(llms.AIChatMessage).ToolCalls {
			toolIDSet[toolCall.ID] = true
			toolCalls = append(toolCalls, toolCall)
		}
		messages = append(messages, message)

		for _, step := range steps {
			toolCallID := step.Action.ToolID
			if ok := toolIDSet[toolCallID]; !ok {
				//don't add tool messages that were not there in previous function call
				continue
			}
			messages = append(messages, llms.ToolChatMessage{
				ID:      toolCallID,
				Content: step.Observation,
			})
		}
	}

	return messages
}

func (o *OpenAIFunctionsAgent) ParseOutput(contentResp *llms.ContentResponse) (
	[]schema.AgentAction, *schema.AgentFinish, []llms.ChatMessage, error,
) {
	var agentActions []schema.AgentAction
	var intermediateMessages []llms.ChatMessage
	for _, choice := range contentResp.Choices {
		// finish
		if len(choice.ToolCalls) == 0 {
			return nil, &schema.AgentFinish{
				ReturnValues: map[string]any{
					"output": choice.Content,
				},
				Log: choice.Content,
			}, nil, nil
		}
		for _, toolCall := range choice.ToolCalls {
			functionName := toolCall.FunctionCall.Name
			toolInputMap := make(map[string]any)
			toolInputStr := toolCall.FunctionCall.Arguments
			err := json.Unmarshal([]byte(toolInputStr), &toolInputMap)
			if err != nil {
				return nil, nil, nil, err
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
			agentActions = append(agentActions, schema.AgentAction{
				Tool:      functionName,
				ToolInput: toolInput,
				Log:       fmt.Sprintf("Invoking: %s with %s \n %s \n", functionName, toolInputStr, contentMsg),
				ToolID:    toolCall.ID,
			})
		}
		intermediateMessages = append(intermediateMessages, choice.ChatMessage)
	}

	return agentActions, nil, intermediateMessages, nil
}
