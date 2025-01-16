package langsmith

import "time"

type TracerSession struct {
	TenantID    string
	ID          string
	StartTime   time.Time
	EndTime     time.Time
	Description *string
	Name        *string
}

type TracerSessionResult struct {
	TracerSession
	RunCount            *int
	LatencyP50          *float64
	LatencyP99          *float64
	TotalTokens         *int
	PromptTokens        *int
	CompletionTokens    *int
	LastRunStartTime    *int64
	FeedbackStats       *KVMap
	ReferenceDatasetIDs *[]string
	RunFacets           *[]KVMap
}

type (
	RunType   string
	ScoreType interface{}
	ValueType interface{}
	DataType  string
)

type BaseExample struct {
	DatasetID string
	Inputs    KVMap
	Outputs   *KVMap
}

type baseRun struct {
	ID                 string   `json:"id,omitempty"`
	Name               string   `json:"name,omitempty"`
	ExecutionOrder     int      `json:"execution_order,omitempty"`
	StartTime          *int64   `json:"start_time,omitempty"`
	RunType            string   `json:"run_type,omitempty"`
	EndTime            *int64   `json:"end_time,omitempty"`
	Extra              KVMap    `json:"extra,omitempty"`
	Error              string   `json:"error,omitempty"`
	Serialized         any      `json:"serialized,omitempty"`
	Events             []KVMap  `json:"events,omitempty"`
	Inputs             KVMap    `json:"inputs,omitempty"`
	Outputs            KVMap    `json:"outputs,omitempty"`
	ReferenceExampleID *string  `json:"reference_example_id,omitempty"`
	ParentRunID        *string  `json:"parent_run_id,omitempty"`
	Tags               []string `json:"tags,omitempty"`
}

type run struct {
	baseRun
	SessionID        *string  `json:"session_id,omitempty"`
	ChildRunIDs      []string `json:"child_run_ids,omitempty"`
	ChildRuns        []run    `json:"child_runs,omitempty"`
	FeedbackStats    KVMap    `json:"feedback_stats,omitempty"`
	AppPath          *string  `json:"app_path,omitempty"`
	ManifestID       *string  `json:"manifest_id,omitempty"`
	Status           *string  `json:"status,omitempty"`
	PromptTokens     *int     `json:"prompt_tokens,omitempty"`
	CompletionTokens *int     `json:"completion_tokens,omitempty"`
	TotalTokens      *int     `json:"total_tokens,omitempty"`
	FirstTokenTime   *int64   `json:"first_token_time,omitempty"`
	ParentRunIDs     []string `json:"parent_run_ids,omitempty"`
	TraceID          *string  `json:"trace_id,omitempty"`
	DottedOrder      *string  `json:"dotted_order,omitempty"`
}

type runCreate struct {
	baseRun
	ChildRuns   []*runCreate `json:"child_runs,omitempty"`
	SessionName *string      `json:"session_name,omitempty"`
}

type runUpdate struct {
	EndTime            *int64  `json:"end_time,omitempty"`
	Extra              KVMap   `json:"extra,omitempty"`
	Error              *string `json:"error,omitempty"`
	Inputs             KVMap   `json:"inputs,omitempty"`
	Outputs            KVMap   `json:"outputs,omitempty"`
	ParentRunID        *string `json:"parent_run_id,omitempty"`
	ReferenceExampleID *string `json:"reference_example_id,omitempty"`
	Events             []KVMap `json:"events,omitempty"`
	SessionID          *string `json:"session_id,omitempty"`
}

type exampleCreate struct {
	BaseExample
	ID        *string
	CreatedAt string
}

type example struct {
	BaseExample
	ID          string
	CreatedAt   string
	ModifiedAt  string
	SourceRunID *string
	Runs        []run
}

type exampleUpdate struct {
	DatasetID *string
	Inputs    *KVMap
	Outputs   *KVMap
}

type baseDataset struct {
	Name     string
	Desc     string
	TenantID string
	DataType *DataType
}

type dataset struct {
	baseDataset
	ID                   string
	CreatedAt            string
	ModifiedAt           string
	ExampleCount         *int
	SessionCount         *int
	LastSessionStartTime *int64
}

type datasetShareSchema struct {
	DatasetID  string
	ShareToken string
	URL        string
}

type feedbackSourceBase struct {
	Type     string
	Metadata *KVMap
}

type apiFeedbackSource struct {
	feedbackSourceBase
}

type modelFeedbackSource struct {
	feedbackSourceBase
}

type feedbackBase struct {
	CreatedAt      time.Time
	ModifiedAt     time.Time
	RunID          string
	Key            string
	Score          ScoreType
	Value          ValueType
	Comment        *string
	Correction     *interface{}
	FeedbackSource *feedbackSourceBase
}

type feedbackCreate struct {
	feedbackBase
	ID string
}

type feedback struct {
	feedbackBase
	ID string
}

type langChainBaseMessage struct {
	GetType          func() string
	Content          string
	AdditionalKwargs *KVMap
}
