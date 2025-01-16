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

var _ callbacks.Handler = (*Tracer)(nil)

type Tracer struct {
	name        string
	projectName string
	client      *Client

	runID      string
	activeTree *runTree
	extras     KVMap
	logger     LeveledLogger
}

func NewTracer(opts ...LangChainTracerOption) (*Tracer, error) {
	tracer := &Tracer{
		name:        "langchain_tracer",
		projectName: "default",
		client:      nil,
		runID:       uuid.New().String(),
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

// GetRunID returns the run ID of the tracer.
func (t *Tracer) GetRunID() string {
	return t.runID
}

func (t *Tracer) resetActiveTree() {
	t.activeTree = nil
}


func (t *Tracer) HandleText(_ context.Context, _ string) {
}

// HandleLLMStart implements callbacks.Handler.
func (t *Tracer) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	childTree := t.activeTree.createChild()

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
		setName("LLMGenerateContent").
		setRunType("llm").
		setInputs(KVMap{
			"messages": inputs,
		})

	t.activeTree.appendChild(childTree)

	// Start the run
	if err := childTree.postRun(ctx, true); err != nil {
		t.logLangSmithError("llm_start", "post run", err)
		return
	}
}

// HandleLLMGenerateContentEnd implements callbacks.Handler.
func (t *Tracer) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	childTree := t.activeTree.getChild("LLMGenerateContent")

	childTree.setName("LLMGenerateContent").setRunType("llm").setOutputs(KVMap{
		"res_content": res,
	})

	// Close the run
	if err := childTree.patchRun(ctx); err != nil {
		t.logLangSmithError("llm_start", "post run", err)
		return
	}
}

// HandleLLMError implements callbacks.Handler.
func (t *Tracer) HandleLLMError(ctx context.Context, err error) {
	t.activeTree.setError(err.Error()).setEndTime(time.Now())

	if err := t.activeTree.patchRun(ctx); err != nil {
		t.logLangSmithError("llm_error", "patch run", err)
		return
	}

	t.activeTree = nil
}

// HandleChainStart implements callbacks.Handler.
func (t *Tracer) HandleChainStart(ctx context.Context, inputs map[string]any) {
	t.activeTree = newRunTree(t.runID).
		setName("RunnableSequence").
		setClient(t.client).
		setProjectName(t.projectName).
		setRunType("chain").
		setInputs(inputs).
		setExtra(t.extras)

	if err := t.activeTree.postRun(ctx, true); err != nil {
		t.logLangSmithError("handle_chain_start", "post run", err)
		return
	}
}

// HandleChainEnd implements callbacks.Handler.
func (t *Tracer) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	t.activeTree.
		setOutputs(outputs).
		setEndTime(time.Now())

	if err := t.activeTree.patchRun(ctx); err != nil {
		t.logLangSmithError("handle_chain_end", "patch run", err)
		return
	}

	t.resetActiveTree()
}

// HandleChainError implements callbacks.Handler.
func (t *Tracer) HandleChainError(ctx context.Context, err error) {
	t.activeTree.setError(err.Error()).setEndTime(time.Now())

	if err := t.activeTree.patchRun(ctx); err != nil {
		t.logLangSmithError("handle_chain_error", "patch run", err)
		return
	}

	t.activeTree = nil
}

// HandleToolStart implements callbacks.Handler.
func (t *Tracer) HandleToolStart(_ context.Context, input string) {
	t.logger.Debugf("handle tool start: input: %s", input)
}

// HandleToolEnd implements callbacks.Handler.
func (t *Tracer) HandleToolEnd(_ context.Context, output string) {
	t.logger.Debugf("handle tool end: output: %s", output)
}

// HandleToolError implements callbacks.Handler.
func (t *Tracer) HandleToolError(_ context.Context, err error) {
	t.logger.Warnf("handle tool error: %s", err)
}

// HandleAgentAction implements callbacks.Handler.
func (t *Tracer) HandleAgentAction(_ context.Context, action schema.AgentAction) {
	t.logger.Debugf("handle agent action, action: %v", action)
}

// HandleAgentFinish implements callbacks.Handler.
func (t *Tracer) HandleAgentFinish(_ context.Context, finish schema.AgentFinish) {
	t.logger.Debugf("handle agent finish, finish: %v", finish)
}

// HandleRetrieverStart implements callbacks.Handler.
func (t *Tracer) HandleRetrieverStart(_ context.Context, query string) {
	t.logger.Debugf("handle retriever start, query: %s, documents: %v", query)
}

// HandleRetrieverEnd implements callbacks.Handler.
func (t *Tracer) HandleRetrieverEnd(_ context.Context, query string, documents []schema.Document) {
	t.logger.Debugf("handle retriever end, query: %s, documents: %v", query, documents)
}

// HandleStreamingFunc implements callbacks.Handler.
func (t *Tracer) HandleStreamingFunc(_ context.Context, _ []byte) {
	// do nothing
}

func (t *Tracer) logLangSmithError(handlerName string, tag string, err error) {
	t.logger.Debugf("we were not able to %s to LangSmith server via handler %q: %s", handlerName, tag, err)
}
