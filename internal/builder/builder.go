package builder

import (
	"context"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/go-faster/errors"
)

type Config struct {
	DockerRegistryPath string
	WasmStoreDir       string
}

type Builder struct {
	cfg Config
}

func New(cfg Config) *Builder {
	return &Builder{
		cfg: cfg,
	}
}

func (b *Builder) Build(ctx context.Context, entry ds.Entry) (ds.Function, error) {
	function := entry.ToFunction()

	switch entry.Runtime {
	case ds.DockerRuntime:
		image, err := b.buildDocker(ctx, entry.UserId, entry.Name, entry.Files)
		if err != nil {
			return ds.Function{}, errors.Wrapf(err, "Build error")
		}
		function.Image = image
	case ds.WasmRuntime:
		wasmPath, err := b.buildWasm(ctx, entry.UserId, entry.Name, entry.Files)
		if err != nil {
			return ds.Function{}, errors.Wrapf(err, "Build error")
		}
		function.WasmPath = wasmPath
	default:
		return ds.Function{}, ds.ErrInvalidRuntime
	}

	return function, nil
}
