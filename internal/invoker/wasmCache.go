package invoker

import (
	"context"
	"os"
	"sync"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/tetratelabs/wazero"
)

type WasmCache struct {
	mu       sync.RWMutex
	compiled map[string]wazero.CompiledModule
}

func NewWasmCache() *WasmCache {
	return &WasmCache{compiled: make(map[string]wazero.CompiledModule)}
}

func (c *WasmCache) GetOrCompile(ctx context.Context, r wazero.Runtime, fn ds.Function) (wazero.CompiledModule, error) {
	c.mu.RLock()
	if m, ok := c.compiled[fn.ID]; ok {
		c.mu.RUnlock()
		return m, nil
	}
	c.mu.RUnlock()

	wasmBytes, err := os.ReadFile(fn.WasmPath)
	if err != nil {
		return nil, err
	}
	compiled, err := r.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.compiled[fn.ID] = compiled
	c.mu.Unlock()
	return compiled, nil
}
