package langsmith

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

type runTree struct {
	ID                  string
	Name                string
	RunType             string
	ProjectName         string
	ParentRun           *runTree
	ChildRuns           []*runTree
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

func newRunTree(id string) *runTree {
	r := runTree{
		ID:                  id,
		ProjectName:         envOr("LANGCHAIN_PROJECT", "default"),
		ExecutionOrder:      1,
		ChildExecutionOrder: 1,
	}
	return r.setStartTime(time.Now())
}

func (t *runTree) setParent(parent *runTree) *runTree {
	t.ParentRun = parent
	return t
}

func (t *runTree) setName(name string) *runTree {
	t.Name = name
	return t
}

func (t *runTree) setProjectName(name string) *runTree {
	t.ProjectName = name
	return t
}

func (t *runTree) setClient(client *Client) *runTree {
	t.Client = client
	return t
}

func (t *runTree) setExecutionOrder(order int) *runTree {
	t.ExecutionOrder = order
	return t
}

func (t *runTree) setChildExecutionOrder(order int) *runTree {
	t.ChildExecutionOrder = order
	return t
}

func (t *runTree) setStartTime(startTime time.Time) *runTree {
	t.StartTime = startTime
	return t
}

func (t *runTree) setEndTime(endTime time.Time) {
	t.EndTime = endTime
}

func (t *runTree) setExtra(extra KVMap) *runTree {
	t.Extra = extra
	return t
}

func (t *runTree) setError(err string) *runTree {
	t.Error = err
	return t
}

func (t *runTree) setInputs(inputs KVMap) *runTree {
	t.Inputs = inputs

	return t
}

func (t *runTree) setOutputs(outputs KVMap) *runTree {
	t.Outputs = outputs
	return t
}

func (t *runTree) setRunType(runType string) *runTree {
	t.RunType = runType
	return t
}

func (t *runTree) createChild() *runTree {
	return newRunTree(uuid.New().String()).
		setParent(t).
		setProjectName(t.ProjectName).
		setClient(t.Client).
		setExecutionOrder(t.ChildExecutionOrder + 1).
		setChildExecutionOrder(t.ChildExecutionOrder + 1)
}

func (t *runTree) appendChild(child *runTree) {
	t.ChildRuns = append(t.ChildRuns, child)
}

func (t *runTree) getChild(childName string) *runTree {
	for _, child := range t.ChildRuns {
		if child.Name == childName {
			return child
		}
	}
	return nil
}

func (t *runTree) convertToCreate(excludeChildRuns bool) (*runCreate, error) {
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

	var childRuns []*runCreate
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
		childRuns = []*runCreate{}
	}

	persistedRun := &runCreate{
		baseRun: baseRun{
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
func (t *runTree) postRun(ctx context.Context, excludeChildRuns bool) error {
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
func (t *runTree) patchRun(ctx context.Context) error {
	var parentRunID *string
	if t.ParentRun != nil {
		parentRunID = valueIfSetOtherwiseNil(t.ParentRun.ID)
	}

	runUpdate := &runUpdate{
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
