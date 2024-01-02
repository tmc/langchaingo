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
	LLM    llms.ChatLLM
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
func NewOpenAIFunctionsAgent(llm llms.ChatLLM, tools []tools.Tool, opts ...CreationOption) *OpenAIFunctionsAgent {
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

	result, err := llms.GenerateChatPrompt(
		ctx,
		o.LLM,
		[]schema.PromptValue{prompt},
		llms.WithFunctions(o.functions()),
		llms.WithStreamingFunc(stream),
	)
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

func createOpenAIFunctionPrompt(opts CreationOptions) prompts.ChatPromptTemplate {
	messageFormatters := []prompts.MessageFormatter{prompts.NewSystemMessagePromptTemplate(opts.systemMessage, nil)}
	messageFormatters = append(messageFormatters, opts.extraMessages...)
	messageFormatters = append(messageFormatters, prompts.NewHumanMessagePromptTemplate("{{.input}}", []string{"input"}))
	messageFormatters = append(messageFormatters, prompts.MessagesPlaceholder{
		VariableName: agentScratchpad,
	})

	tmpl := prompts.NewChatPromptTemplate(messageFormatters)
	return tmpl
}

func (o *OpenAIFunctionsAgent) constructScratchPad(steps []schema.AgentStep) []schema.ChatMessage {
	if len(steps) == 0 {
		return nil
	}

	messages := make([]schema.ChatMessage, 0)
	for _, step := range steps {
		messages = append(messages, schema.FunctionChatMessage{
			Name:    step.Action.Tool,
			Content: step.Observation,
		})
	}

	return messages
}

func (o *OpenAIFunctionsAgent) ParseOutput(llmResult llms.LLMResult) (
	[]schema.AgentAction, *schema.AgentFinish, error,
) {
	if llmResult.Generations[0][0] == nil || llmResult.Generations[0][0].Message == nil {
		return nil, nil, fmt.Errorf("%w: %s", ErrInvalidLLMGenerationsMessageReturn,
			"wrong llmResult from OpenAIFunctionsAgent")
	}

	// finish
	if llmResult.Generations[0][0].Message.FunctionCall == nil {
		return nil, &schema.AgentFinish{
			ReturnValues: map[string]any{
				"output": llmResult.Generations[0][0].Message.Content,
			},
			Log: llmResult.Generations[0][0].Message.Content,
		}, nil
	}

	// action
	functionCall := llmResult.Generations[0][0].Message.FunctionCall
	functionName := functionCall.Name
	toolInputStr := functionCall.Arguments
	toolInputMap := make(map[string]any, 0)
	err := json.Unmarshal([]byte(toolInputStr), &toolInputMap)
	if err != nil {
		return nil, nil, err
	}

	toolInput := toolInputStr
	if arg1, ok := toolInputMap["__arg1"]; ok {
		toolInputCheck, ok := arg1.(string)
		if ok {
			toolInput = toolInputCheck
		}
	}

	contentMsg := "\n"
	if llmResult.Generations[0][0].Message.Content != "" {
		contentMsg = fmt.Sprintf("responded: %s\n", llmResult.Generations[0][0].Message.Content)
	}

	return []schema.AgentAction{
		{
			Tool:      functionName,
			ToolInput: toolInput,
			Log:       fmt.Sprintf("Invoking: %s with %s \n %s \n", functionName, toolInputStr, contentMsg),
		},
	}, nil, nil
}
