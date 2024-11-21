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

func (t *RunTree) SetError(error string) *RunTree {
	t.Error = error
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

func (r *RunTree) CreateChild() *RunTree {
	return NewRunTree(uuid.New().String()).
		SetParent(r).
		SetProjectName(r.ProjectName).
		SetClient(r.Client).
		SetExecutionOrder(r.ChildExecutionOrder + 1).
		SetChildExecutionOrder(r.ChildExecutionOrder + 1)
}

func (r *RunTree) AppendChild(child *RunTree) {
	r.ChildRuns = append(r.ChildRuns, child)
}

func (t *RunTree) GetChild(childName string) *RunTree {
	for _, child := range t.ChildRuns {
		if child.Name == childName {
			return child
		}
	}
	return nil
}

func (r *RunTree) End(outputs KVMap, error string, endTime time.Time) {
	r.Outputs = outputs
	r.Error = error
	r.EndTime = endTime

	if r.ParentRun != nil {
		r.ParentRun.ChildExecutionOrder = max(
			r.ParentRun.ChildExecutionOrder,
			r.ChildExecutionOrder,
		)
	}
}
func (r *RunTree) convertToCreate(excludeChildRuns bool) (*RunCreate, error) {
	runExtra := r.Extra
	if runExtra == nil {
		runExtra = make(KVMap)
	}

	var runExtraRuntime KVMap
	if v := runExtra["runtime"]; v == nil {
		runExtraRuntime = make(KVMap)
		runExtra["runtime"] = runExtraRuntime
	} else {
		runExtraRuntime = v.(KVMap)
	}

	runExtraRuntime["library"] = "langsmith-go"
	runExtraRuntime["commit"] = getGitCommit()

	var childRuns []*RunCreate
	var parentRunID *string
	if !excludeChildRuns {
		for _, childRun := range r.ChildRuns {
			childRunCreate, err := childRun.convertToCreate(excludeChildRuns)
			if err != nil {
				return nil, err
			}
			childRuns = append(childRuns, childRunCreate)
		}
		parentRunID = nil
	} else {
		if r.ParentRun != nil {
			parentRunID = &r.ParentRun.ID
		}
		childRuns = []*RunCreate{}
	}

	persistedRun := &RunCreate{
		BaseRun: BaseRun{
			ID:                 r.ID,
			Name:               r.Name,
			StartTime:          timeToMillisecondsPtr(r.StartTime),
			EndTime:            timeToMillisecondsPtr(r.EndTime),
			RunType:            r.RunType,
			ReferenceExampleID: r.ReferenceExampleID,
			Extra:              runExtra,
			ExecutionOrder:     r.ExecutionOrder,
			Serialized:         r.Serialized,
			Error:              r.Error,
			Inputs:             r.Inputs,
			Outputs:            r.Outputs,
			ParentRunID:        parentRunID,
		},
		SessionName: &r.ProjectName,
		ChildRuns:   childRuns,
	}
	return persistedRun, nil
}

// postRun will start the run or the child run
func (r *RunTree) postRun(ctx context.Context, excludeChildRuns bool) error {
	runCreate, err := r.convertToCreate(true)
	if err != nil {
		return err
	}

	err = r.Client.CreateRun(ctx, runCreate)
	if err != nil {
		return err
	}

	if !excludeChildRuns {
		for _, childRun := range r.ChildRuns {
			err = childRun.postRun(ctx, excludeChildRuns)
			if err != nil {
				return fmt.Errorf("post child: %w", err)
			}
		}
	}

	return nil
}

// patchRun will close the run or the child run of a run tree that
// was started with the postRun
func (r *RunTree) patchRun(ctx context.Context) error {
	var parentRunID *string
	if r.ParentRun != nil {
		parentRunID = valueIfSetOtherwiseNil(r.ParentRun.ID)
	}

	runUpdate := &RunUpdate{
		EndTime:            timeToMillisecondsPtr(r.EndTime),
		Error:              valueIfSetOtherwiseNil(r.Error),
		Outputs:            r.Outputs,
		ParentRunID:        parentRunID,
		ReferenceExampleID: r.ReferenceExampleID,
		Extra:              r.Extra,
		Events:             r.Events,
	}

	err := r.Client.UpdateRun(ctx, r.ID, runUpdate)
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

	findSetting := func(key string, settings []debug.BuildSetting) (value string) {
		for _, setting := range settings {
			if setting.Key == key {
				return setting.Value
			}
		}

		return ""
	}

	return findSetting("vcs.revision", info.Settings)
}
