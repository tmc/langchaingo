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
