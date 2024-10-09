package scheduler

import (
	"testing"
	"time"

	"github.com/helixml/helix/api/pkg/model"
	"github.com/helixml/helix/api/pkg/types"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

const (
	testModelStr = types.Model_Ollama_Llama3_8b
)

var (
	dummyTimeout = func(runnerID string, lastActivityTime time.Time) bool {
		return false
	}
	testModel, _ = model.GetModel(testModelStr)
)

func TestPlacement_MaxSpread_Simple(t *testing.T) {
	c := NewCluster(dummyTimeout)
	c.UpdateRunner(&types.RunnerState{
		ID:          "test-runner-1",
		TotalMemory: testModel.GetMemoryRequirements(types.SessionModeInference) * 2,
	})
	a := NewWorkloadAllocator(dummyTimeout)

	req := createPlacementWork("test", testModelStr)

	runnerID, err := MaxSpreadStrategy(c, a, req)
	assert.NoError(t, err)
	assert.Equal(t, "test-runner-1", runnerID)
	a.AllocateNewSlot(runnerID, req)

	runnerID, err = MaxSpreadStrategy(c, a, req)
	assert.NoError(t, err)
	assert.Equal(t, "test-runner-1", runnerID)
}

func TestPlacement_MaxSpread_MultiMachine(t *testing.T) {
	c := NewCluster(dummyTimeout)
	c.UpdateRunner(&types.RunnerState{
		ID:          "test-runner-1",
		TotalMemory: 2 * testModel.GetMemoryRequirements(types.SessionModeInference),
	})
	a := NewWorkloadAllocator(dummyTimeout)
	req := createPlacementWork("test", testModelStr)
	a.AllocateNewSlot("test-runner-1", req)

	// Add a second runner
	c.UpdateRunner(&types.RunnerState{
		ID:          "test-runner-2",
		TotalMemory: 2 * testModel.GetMemoryRequirements(types.SessionModeInference),
	})

	runnerID, err := MaxSpreadStrategy(c, a, req)
	assert.NoError(t, err)
	assert.Equal(t, "test-runner-2", runnerID)
}

func createPlacementWork(name string, model types.ModelName) *Workload {
	req := &types.RunnerLLMInferenceRequest{
		RequestID: name,
		Request: &openai.ChatCompletionRequest{
			Model: model.String(),
		},
	}
	work, _ := NewLLMWorkload(req)
	return work
}