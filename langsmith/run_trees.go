package langsmith

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

type RunTree struct {
	ID                  string
	Name                string
	RunType             string
	ProjectName         string
	ParentRun           *RunTree
	ChildRuns           []*RunTree
	ExecutionOrder      int
	ChildExecutionOrder int
	StartTime           time.Time
	EndTime             time.Time
	Extra               KVMap
	Error               string
	Serialized          KVMap
	Inputs              KVMap
	Outputs             KVMap
	ReferenceExampleID  *string
	Client              *Client
	Events              []KVMap
}

func NewRunTree(id string) *RunTree {
	return &RunTree{
		ID:                  id,
		ProjectName:         envOr("LANGCHAIN_PROJECT", "default"),
		ExecutionOrder:      1,
		ChildExecutionOrder: 1,
		StartTime:           time.Now(),
	}
}

func (t *RunTree) SetParent(parent *RunTree) *RunTree {
	t.ParentRun = parent
	return t
}

func (t *RunTree) SetName(name string) *RunTree {
	t.Name = name
	return t
}

func (t *RunTree) SetProjectName(name string) *RunTree {
	t.ProjectName = name
	return t
}

func (t *RunTree) SetClient(client *Client) *RunTree {
	t.Client = client
	return t
}

func (t *RunTree) SetExecutionOrder(order int) *RunTree {
	t.ExecutionOrder = order
	return t
}

func (t *RunTree) SetChildExecutionOrder(order int) *RunTree {
	t.ChildExecutionOrder = order
	return t
}

func (t *RunTree) SetStartTime(startTime time.Time) *RunTree {
	t.StartTime = startTime
	return t
}

func (t *RunTree) SetEndTime(endTime time.Time) *RunTree {
	t.EndTime = endTime
	return t
}

func (t *RunTree) SetExtra(extra KVMap) *RunTree {
	t.Extra = extra
	return t
}

func (t *RunTree) SetError(err string) *RunTree {
	t.Error = err
	return t
}

func (t *RunTree) SetSerialized(serialized KVMap) *RunTree {
	t.Serialized = serialized
	return t
}

func (t *RunTree) SetInputs(inputs KVMap) *RunTree {
	t.Inputs = inputs

	return t
}

func (t *RunTree) SetOutputs(outputs KVMap) *RunTree {
	t.Outputs = outputs
	return t
}

func (t *RunTree) SetReferenceExampleID(id string) *RunTree {
	t.ReferenceExampleID = &id
	return t
}

func (t *RunTree) SetEvents(events []KVMap) *RunTree {
	t.Events = events
	return t
}

func (t *RunTree) SetRunType(runType string) *RunTree {
	t.RunType = runType
	return t
}

func (t *RunTree) CreateChild() *RunTree {
	return NewRunTree(uuid.New().String()).
		SetParent(t).
		SetProjectName(t.ProjectName).
		SetClient(t.Client).
		SetExecutionOrder(t.ChildExecutionOrder + 1).
		SetChildExecutionOrder(t.ChildExecutionOrder + 1)
}

func (t *RunTree) AppendChild(child *RunTree) {
	t.ChildRuns = append(t.ChildRuns, child)
}

func (t *RunTree) GetChild(childName string) *RunTree {
	for _, child := range t.ChildRuns {
		if child.Name == childName {
			return child
		}
	}
	return nil
}

func (t *RunTree) End(outputs KVMap, err string, endTime time.Time) {
	t.Outputs = outputs
	t.Error = err
	t.EndTime = endTime

	if t.ParentRun != nil {
		t.ParentRun.ChildExecutionOrder = max(
			t.ParentRun.ChildExecutionOrder,
			t.ChildExecutionOrder,
		)
	}
}

func (t *RunTree) convertToCreate(excludeChildRuns bool) (*RunCreate, error) {
	runExtra := t.Extra
	if runExtra == nil {
		runExtra = make(KVMap)
	}

	var runExtraRuntime KVMap
	if v := runExtra["runtime"]; v == nil {
		runExtraRuntime = make(KVMap)
		runExtra["runtime"] = runExtraRuntime
	} else if runExtraRuntimeCast, ok := v.(KVMap); ok {
		runExtraRuntime = runExtraRuntimeCast
	} else {
		return nil, fmt.Errorf("extra must be a map")
	}

	runExtraRuntime["library"] = "langsmith-go"
	runExtraRuntime["commit"] = getGitCommit()

	var childRuns []*RunCreate
	var parentRunID *string
	if !excludeChildRuns {
		for _, childRun := range t.ChildRuns {
			childRunCreate, err := childRun.convertToCreate(excludeChildRuns)
			if err != nil {
				return nil, err
			}
			childRuns = append(childRuns, childRunCreate)
		}
		parentRunID = nil
	} else {
		if t.ParentRun != nil {
			parentRunID = &t.ParentRun.ID
		}
		childRuns = []*RunCreate{}
	}

	persistedRun := &RunCreate{
		BaseRun: BaseRun{
			ID:                 t.ID,
			Name:               t.Name,
			StartTime:          timeToMillisecondsPtr(t.StartTime),
			EndTime:            timeToMillisecondsPtr(t.EndTime),
			RunType:            t.RunType,
			ReferenceExampleID: t.ReferenceExampleID,
			Extra:              runExtra,
			ExecutionOrder:     t.ExecutionOrder,
			Serialized:         t.Serialized,
			Error:              t.Error,
			Inputs:             t.Inputs,
			Outputs:            t.Outputs,
			ParentRunID:        parentRunID,
		},
		SessionName: &t.ProjectName,
		ChildRuns:   childRuns,
	}
	return persistedRun, nil
}

// postRun will start the run or the child run.
func (t *RunTree) postRun(ctx context.Context, excludeChildRuns bool) error {
	runCreate, err := t.convertToCreate(true)
	if err != nil {
		return err
	}

	err = t.Client.CreateRun(ctx, runCreate)
	if err != nil {
		return err
	}

	if !excludeChildRuns {
		for _, childRun := range t.ChildRuns {
			err = childRun.postRun(ctx, excludeChildRuns)
			if err != nil {
				return fmt.Errorf("post child: %w", err)
			}
		}
	}

	return nil
}

// patchRun will close the run or the child run of a run tree that
// was started with the postRun.
func (t *RunTree) patchRun(ctx context.Context) error {
	var parentRunID *string
	if t.ParentRun != nil {
		parentRunID = valueIfSetOtherwiseNil(t.ParentRun.ID)
	}

	runUpdate := &RunUpdate{
		EndTime:            timeToMillisecondsPtr(t.EndTime),
		Error:              valueIfSetOtherwiseNil(t.Error),
		Outputs:            t.Outputs,
		ParentRunID:        parentRunID,
		ReferenceExampleID: t.ReferenceExampleID,
		Extra:              t.Extra,
		Events:             t.Events,
	}

	err := t.Client.UpdateRun(ctx, t.ID, runUpdate)
	if err != nil {
		return err
	}

	return nil
}

func getGitCommit() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("we should have been able to retrieve info from 'runtime/debug#ReadBuildInfo'")
	}

	findSetting := func(key string, settings []debug.BuildSetting) string {
		for _, setting := range settings {
			if setting.Key == key {
				return setting.Value
			}
		}

		return ""
	}

	return findSetting("vcs.revision", info.Settings)
}
