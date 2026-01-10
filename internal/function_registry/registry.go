package functionregistry

import (
	"context"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/go-faster/errors"
)

type repository interface {
	SaveFunction(ctx context.Context, function ds.Function) (id string, err error)
	GetFunction(ctx context.Context, id string) (function ds.Function, err error)
}

type builder interface {
	Build(ctx context.Context, name string, files map[string][]byte) (string, error)
}

type Registry struct {
	repo    repository
	builder builder
}

func New(repo repository, builder builder) *Registry {
	return &Registry{
		repo:    repo,
		builder: builder,
	}
}

func (r *Registry) Register(ctx context.Context, entry ds.Entry) (id string, err error) {
	function := entry.ToFunction()

	switch entry.Runtime {
	case ds.DockerRuntime:
		image, err := r.builder.Build(ctx, entry.Name, entry.Files)
		if err != nil {
			return "", errors.Wrapf(err, "Build error")
		}
		function.Image = image
	case ds.WasmRuntime:
		// build wasm
	default:
		return "", ds.ErrInvalidRuntime
	}

	id, err = r.repo.SaveFunction(ctx, function)
	if err != nil {
		return "", errors.Wrapf(err, "Save function error")
	}

	return id, nil

}

func (r *Registry) Get(ctx context.Context, id string) (ds.Function, error) {
	return r.repo.GetFunction(ctx, id)
}
