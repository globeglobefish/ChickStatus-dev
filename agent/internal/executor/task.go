package executor

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/probe-system/agent/pkg/protocol"
)

type TaskManager struct {
	pingExecutor   *PingExecutor
	scriptExecutor *ScriptExecutor
	tasks          sync.Map // taskID -> *RunningTask
	resultChan     chan *protocol.TaskResultPayload
	stopChan       chan struct{}
}

type RunningTask struct {
	ID       string
	Type     string
	Cancel   context.CancelFunc
	Interval int
}

func NewTaskManager(coreURL, scriptDir string) *TaskManager {
	return &TaskManager{
		pingExecutor:   NewPingExecutor(5*time.Second, 4),
		scriptExecutor: NewScriptExecutor(coreURL, scriptDir, 60*time.Second),
		resultChan:     make(chan *protocol.TaskResultPayload, 100),
		stopChan:       make(chan struct{}),
	}
}

func (m *TaskManager) GetResultChan() <-chan *protocol.TaskResultPayload {
	return m.resultChan
}

func (m *TaskManager) HandleTask(task *protocol.TaskAssignPayload) {
	// Cancel existing task with same ID
	if existing, ok := m.tasks.Load(task.TaskID); ok {
		rt := existing.(*RunningTask)
		rt.Cancel()
		m.tasks.Delete(task.TaskID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	rt := &RunningTask{
		ID:       task.TaskID,
		Type:     task.Type,
		Cancel:   cancel,
		Interval: task.Interval,
	}
	m.tasks.Store(task.TaskID, rt)

	go m.runTask(ctx, task)
}

func (m *TaskManager) runTask(ctx context.Context, task *protocol.TaskAssignPayload) {
	defer func() {
		if task.Interval <= 0 {
			m.tasks.Delete(task.TaskID)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		default:
		}

		start := time.Now()
		result := &protocol.TaskResultPayload{
			TaskID: task.TaskID,
		}

		switch task.Type {
		case "ping":
			pingResult, err := m.pingExecutor.Execute(ctx, task.Target)
			if err != nil {
				result.Success = false
				result.Error = err.Error()
			} else {
				result.Success = pingResult.Success
				output, _ := json.Marshal(pingResult)
				result.Output = string(output)
				if !pingResult.Success {
					result.Error = pingResult.Error
				}
			}

		case "script":
			checksum := ""
			if task.Params != nil {
				checksum = task.Params["checksum"]
			}
			scriptResult, err := m.scriptExecutor.Execute(ctx, task.ScriptID, task.Params, checksum)
			if err != nil {
				result.Success = false
				result.Error = err.Error()
			} else {
				result.Success = scriptResult.ExitCode == 0
				result.Output = scriptResult.Stdout
				if scriptResult.Stderr != "" {
					result.Error = scriptResult.Stderr
				}
			}

		default:
			result.Success = false
			result.Error = "unknown task type"
		}

		result.Duration = time.Since(start).Milliseconds()

		select {
		case m.resultChan <- result:
		default:
			log.Printf("Result channel full, dropping result for task %s", task.TaskID)
		}

		// If not recurring, exit
		if task.Interval <= 0 {
			return
		}

		// Wait for next interval
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-time.After(time.Duration(task.Interval) * time.Second):
		}
	}
}

func (m *TaskManager) CancelTask(taskID string) {
	if existing, ok := m.tasks.Load(taskID); ok {
		rt := existing.(*RunningTask)
		rt.Cancel()
		m.tasks.Delete(taskID)
	}
}

func (m *TaskManager) Stop() {
	close(m.stopChan)
	m.tasks.Range(func(key, value interface{}) bool {
		rt := value.(*RunningTask)
		rt.Cancel()
		return true
	})
}
