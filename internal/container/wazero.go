package container

import (
	"context"
	"os"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/tetratelabs/wazero"
)

func (c *Container) NewWasmRuntime(ctx context.Context) (wazero.Runtime, error) {
	if err := os.MkdirAll(ds.WasmCacheDir, 0755); err != nil {
		return nil, err
	}

	compCache, err := wazero.NewCompilationCacheWithDir(ds.WasmCacheDir)
	if err != nil {
		return nil, err
	}

	config := wazero.NewRuntimeConfig().
		WithCompilationCache(compCache).
		WithMemoryLimitPages(c.cfg.Wasm.MaxPages).
		WithCloseOnContextDone(true)

	return wazero.NewRuntimeWithConfig(ctx, config), nil
}
