package pool

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockTask 模拟任务
type MockTask struct {
	id       string
	duration time.Duration
	err      error
}

func (t *MockTask) Execute(ctx context.Context) (interface{}, error) {
	if t.duration > 0 {
		time.Sleep(t.duration)
	}

	if t.err != nil {
		return nil, t.err
	}

	return "result-" + t.id, nil
}

func (t *MockTask) ID() string {
	return t.id
}

func TestWorkerPool_StartStop(t *testing.T) {
	pool := NewWorkerPool(2, 10)

	if pool.IsStarted() {
		t.Error("Pool should not be started initially")
	}

	err := pool.Start()
	if err != nil {
		t.Fatalf("Failed to start pool: %v", err)
	}

	if !pool.IsStarted() {
		t.Error("Pool should be started")
	}

	pool.Stop()

	if pool.IsStarted() {
		t.Error("Pool should be stopped")
	}
}

func TestWorkerPool_Submit(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	pool.Start()
	defer pool.Stop()

	task := &MockTask{
		id:       "task1",
		duration: time.Millisecond * 10,
	}

	err := pool.Submit(task)
	if err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	// 等待结果
	select {
	case result := <-pool.Results():
		if result.TaskID != "task1" {
			t.Errorf("Expected task1, got %s", result.TaskID)
		}
		if result.Error != nil {
			t.Errorf("Unexpected error: %v", result.Error)
		}
		if result.Value != "result-task1" {
			t.Errorf("Expected result-task1, got %v", result.Value)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for result")
	}
}

func TestWorkerPool_Error(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	pool.Start()
	defer pool.Stop()

	task := &MockTask{
		id:  "task-error",
		err: errors.New("test error"),
	}

	pool.Submit(task)

	select {
	case result := <-pool.Results():
		if result.Error == nil {
			t.Error("Expected error")
		}
		if result.Error.Error() != "test error" {
			t.Errorf("Expected 'test error', got %v", result.Error)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for result")
	}
}

func TestWorkerPool_Multiple(t *testing.T) {
	pool := NewWorkerPool(4, 20)
	pool.Start()
	defer pool.Stop()

	taskCount := 10
	for i := 0; i < taskCount; i++ {
		task := &MockTask{
			id:       string(rune('0' + i)),
			duration: time.Millisecond * 10,
		}
		pool.Submit(task)
	}

	// 收集结果
	results := make(map[string]bool)
	for i := 0; i < taskCount; i++ {
		select {
		case result := <-pool.Results():
			results[result.TaskID] = true
		case <-time.After(time.Second * 2):
			t.Fatal("Timeout waiting for results")
		}
	}

	if len(results) != taskCount {
		t.Errorf("Expected %d results, got %d", taskCount, len(results))
	}
}