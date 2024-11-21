package langsmith

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

var _ callbacks.Handler = (*LangChainTracer)(nil)

type LangChainTracer struct {
	name        string
	projectName string
	client      *Client

	runId      string
	activeTree *RunTree
	extras     KVMap
	logger     LeveledLoggerInterface
}

func (t *LangChainTracer) HandleLLMStart(ctx context.Context, prompts []string) {}

func NewTracer(opts ...LangChainTracerOption) (*LangChainTracer, error) {
	tracer := &LangChainTracer{
		name:        "langchain_tracer",
		projectName: "default",
		client:      nil,
		runId:       uuid.New().String(),
		logger:      &NopLogger{},
	}

	for _, opt := range opts {
		opt.apply(tracer)
	}

	if tracer.client == nil {
		var err error
		tracer.client, err = NewClient()
		if err != nil {
			return nil, fmt.Errorf("new langsmith client: %w", err)
		}
	}

	return tracer, nil
}

func (t *LangChainTracer) GetRunID() string {
	return t.runId
}

func (t *LangChainTracer) resetActiveTree() {
	t.activeTree = nil
}

// HandleText implements callbacks.Handler.
func (t *LangChainTracer) HandleText(ctx context.Context, text string) {
}

func (t *LangChainTracer) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	childTree := t.activeTree.CreateChild()

	inputs := []struct {
		Role    string             `json:"role"`
		Content []llms.ContentPart `json:"content"`
	}{}

	for _, prompt := range ms {
		inputs = append(inputs, struct {
			Role    string             `json:"role"`
			Content []llms.ContentPart `json:"content"`
		}{
			Role:    string(prompt.Role),
			Content: prompt.Parts,
		})
	}

	childTree.
		SetName("LLMGenerateContent").
		SetRunType("llm").
		SetInputs(KVMap{
			"messages": inputs,
		})

	t.activeTree.AppendChild(childTree)

	// Start the run
	if err := childTree.postRun(ctx, true); err != nil {
		t.logLangSmithError("llm_start", "post run", err)
		return
	}
}

func (t *LangChainTracer) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	childTree := t.activeTree.GetChild("LLMGenerateContent")

	childTree.SetName("LLMGenerateContent").SetRunType("llm").SetOutputs(KVMap{
		"res_content": res,
	})

	// Close the run
	if err := childTree.patchRun(ctx); err != nil {
		t.logLangSmithError("llm_start", "post run", err)
		return
	}
}

// HandleLLMError implements callbacks.Handler.
func (t *LangChainTracer) HandleLLMError(ctx context.Context, err error) {
	t.activeTree.SetError(err.Error()).SetEndTime(time.Now())

	if err := t.activeTree.patchRun(ctx); err != nil {
		t.logLangSmithError("llm_error", "patch run", err)
		return
	}

	t.activeTree = nil
}

// HandleChainStart implements callbacks.Handler.
func (t *LangChainTracer) HandleChainStart(ctx context.Context, inputs map[string]any) {
	t.activeTree = NewRunTree(t.runId).
		SetName("RunnableSequence").
		SetClient(t.client).
		SetProjectName(t.projectName).
		SetRunType("chain").
		SetInputs(inputs).
		SetExtra(t.extras)

	if err := t.activeTree.postRun(ctx, true); err != nil {
		t.logLangSmithError("handle_chain_start", "post run", err)
		return
	}
}

// HandleChainEnd implements callbacks.Handler.
func (t *LangChainTracer) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	t.activeTree.
		SetOutputs(outputs).
		SetEndTime(time.Now())

	if err := t.activeTree.patchRun(ctx); err != nil {
		t.logLangSmithError("handle_chain_end", "patch run", err)
		return
	}

	t.resetActiveTree()
}

// HandleChainError implements callbacks.Handler.
func (t *LangChainTracer) HandleChainError(ctx context.Context, err error) {
	t.activeTree.SetError(err.Error()).SetEndTime(time.Now())

	if err := t.activeTree.patchRun(ctx); err != nil {
		t.logLangSmithError("handle_chain_error", "patch run", err)
		return
	}

	t.activeTree = nil
}

// HandleToolStart implements callbacks.Handler.
func (t *LangChainTracer) HandleToolStart(ctx context.Context, input string) {
	t.logger.Debugf("handle tool start: input: %s", input)
}

// HandleToolEnd implements callbacks.Handler.
func (t *LangChainTracer) HandleToolEnd(ctx context.Context, output string) {
	t.logger.Debugf("handle tool end: output: %s", output)
}

// HandleToolError implements callbacks.Handler.
func (t *LangChainTracer) HandleToolError(ctx context.Context, err error) {
	t.logger.Warnf("handle tool error: %s", err)
}

// HandleAgentAction implements callbacks.Handler.
func (t *LangChainTracer) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	t.logger.Debugf("handle agent action, action: %v", action)
}

// HandleAgentFinish implements callbacks.Handler.
func (t *LangChainTracer) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	t.logger.Debugf("handle agent finish, finish: %v", finish)

}

// HandleRetrieverStart implements callbacks.Handler.
func (t *LangChainTracer) HandleRetrieverStart(ctx context.Context, query string) {
	t.logger.Debugf("handle retriever start, query: %s, documents: %v", query)
}

// HandleRetrieverEnd implements callbacks.Handler.
func (t *LangChainTracer) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	t.logger.Debugf("handle retriever end, query: %s, documents: %v", query, documents)
}

// HandleStreamingFunc implements callbacks.Handler.
func (t *LangChainTracer) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	// do nothing
}

func (t *LangChainTracer) logLangSmithError(handlerName string, tag string, err error) {
	t.logger.Debugf("we were not able to %s to LangSmith server via handler %q: %s", handlerName, tag, err)
}
