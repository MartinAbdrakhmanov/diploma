package container

import (
	"context"
	"os"

	"github.com/tetratelabs/wazero"
)

func (c *Container) NewWasmRuntime(ctx context.Context) (wazero.Runtime, error) {
	cacheDir := "/test/wasm_cache_v1" //TODO change me
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	compCache, err := wazero.NewCompilationCacheWithDir(cacheDir)
	if err != nil {
		return nil, err
	}

	config := wazero.NewRuntimeConfig().WithCompilationCache(compCache).WithMemoryLimitPages(c.cfg.Wasm.MaxPages)

	return wazero.NewRuntimeWithConfig(ctx, config), nil
}
