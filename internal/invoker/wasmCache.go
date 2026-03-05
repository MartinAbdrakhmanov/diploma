package invoker

import (
	"context"
	"os"
	"sync"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/tetratelabs/wazero"
	"golang.org/x/sync/singleflight"
)

// TODO delete/change, now wazero cache is used
type WasmCache struct {
	mu       sync.RWMutex
	compiled map[string]wazero.CompiledModule
	sf       singleflight.Group
}

func NewWasmCache() *WasmCache {
	return &WasmCache{
		compiled: make(map[string]wazero.CompiledModule),
	}
}

func (c *WasmCache) GetOrCompile(ctx context.Context, r wazero.Runtime, fn ds.Function) (wazero.CompiledModule, error) {
	c.mu.RLock()
	if m, ok := c.compiled[fn.ID]; ok {
		c.mu.RUnlock()
		return m, nil
	}
	c.mu.RUnlock()

	val, err, _ := c.sf.Do(fn.ID, func() (interface{}, error) {
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
	})

	if err != nil {
		return nil, err
	}

	return val.(wazero.CompiledModule), nil
}
