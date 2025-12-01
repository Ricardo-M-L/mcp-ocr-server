package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ricardo/mcp-ocr-server/pkg/logger"
	"go.uber.org/zap"
)

// Task 任务接口
type Task interface {
	Execute(ctx context.Context) (interface{}, error)
	ID() string
}

// Result 任务结果
type Result struct {
	TaskID string
	Value  interface{}
	Error  error
	Duration time.Duration
}

// WorkerPool Worker 池
type WorkerPool struct {
	workerCount int
	taskQueue   chan Task
	resultQueue chan *Result
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	started     bool
	mu          sync.RWMutex
}

// NewWorkerPool 创建 Worker 池
func NewWorkerPool(workerCount, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan Task, queueSize),
		resultQueue: make(chan *Result, queueSize),
		ctx:         ctx,
		cancel:      cancel,
		started:     false,
	}
}

// Start 启动 Worker 池
func (p *WorkerPool) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return fmt.Errorf("worker pool already started")
	}

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	p.started = true
	logger.Info("Worker pool started", zap.Int("workers", p.workerCount))

	return nil
}

// Stop 停止 Worker 池
func (p *WorkerPool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return
	}

	// 关闭任务队列
	close(p.taskQueue)

	// 等待所有 worker 完成
	p.wg.Wait()

	// 取消上下文
	p.cancel()

	// 关闭结果队列
	close(p.resultQueue)

	p.started = false
	logger.Info("Worker pool stopped")
}

// Submit 提交任务
func (p *WorkerPool) Submit(task Task) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started {
		return fmt.Errorf("worker pool not started")
	}

	select {
	case p.taskQueue <- task:
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("worker pool stopped")
	default:
		return fmt.Errorf("task queue full")
	}
}

// Results 获取结果通道
func (p *WorkerPool) Results() <-chan *Result {
	return p.resultQueue
}

// worker Worker 工作协程
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	logger.Debug("Worker started", zap.Int("worker_id", id))

	for {
		select {
		case task, ok := <-p.taskQueue:
			if !ok {
				logger.Debug("Worker stopped", zap.Int("worker_id", id))
				return
			}

			// 执行任务
			startTime := time.Now()
			value, err := task.Execute(p.ctx)
			duration := time.Since(startTime)

			result := &Result{
				TaskID:   task.ID(),
				Value:    value,
				Error:    err,
				Duration: duration,
			}

			// 发送结果
			select {
			case p.resultQueue <- result:
			case <-p.ctx.Done():
				return
			}

		case <-p.ctx.Done():
			logger.Debug("Worker stopped", zap.Int("worker_id", id))
			return
		}
	}
}

// Stats 获取统计信息
func (p *WorkerPool) Stats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"worker_count": p.workerCount,
		"queue_size":   len(p.taskQueue),
		"queue_cap":    cap(p.taskQueue),
		"started":      p.started,
	}
}

// IsStarted 检查是否已启动
func (p *WorkerPool) IsStarted() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.started
}

// QueueSize 获取队列大小
func (p *WorkerPool) QueueSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.taskQueue)
}