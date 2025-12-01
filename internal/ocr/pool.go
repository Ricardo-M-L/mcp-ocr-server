package ocr

import (
	"fmt"
	"sync"
)

// EnginePool 引擎资源池
type EnginePool struct {
	cgoEngines chan Engine
	config     EngineConfig
	mu         sync.Mutex
}

// NewEnginePool 创建新的引擎池
func NewEnginePool(config EngineConfig, poolSize int) (*EnginePool, error) {
	pool := &EnginePool{
		cgoEngines: make(chan Engine, poolSize),
		config:     config,
	}

	// 预创建引擎实例
	for i := 0; i < poolSize; i++ {
		engine := NewTesseractCGoEngine()
		if err := engine.Initialize(config); err != nil {
			// 关闭已创建的引擎
			close(pool.cgoEngines)
			for e := range pool.cgoEngines {
				e.Close()
			}
			return nil, fmt.Errorf("failed to initialize engine %d: %w", i, err)
		}
		pool.cgoEngines <- engine
	}

	return pool, nil
}

// Get 从池中获取引擎
func (p *EnginePool) Get() (Engine, error) {
	select {
	case engine := <-p.cgoEngines:
		return engine, nil
	default:
		// 池为空，创建新引擎
		engine := NewTesseractCGoEngine()
		if err := engine.Initialize(p.config); err != nil {
			return nil, fmt.Errorf("failed to create new engine: %w", err)
		}
		return engine, nil
	}
}

// Put 将引擎返回池中
func (p *EnginePool) Put(engine Engine) {
	if engine == nil {
		return
	}

	select {
	case p.cgoEngines <- engine:
		// 成功放回池中
	default:
		// 池已满，关闭引擎
		engine.Close()
	}
}

// Close 关闭引擎池
func (p *EnginePool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.cgoEngines)

	var lastErr error
	for engine := range p.cgoEngines {
		if err := engine.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}