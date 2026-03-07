package functionregistry

import (
	"context"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/go-faster/errors"
)

type repository interface {
	SaveFunction(ctx context.Context, function ds.Function) (id string, err error)
	GetFunction(ctx context.Context, id string) (function ds.Function, err error)
	DeleteFunction(ctx context.Context, id string) error
}

type builder interface {
	Build(ctx context.Context, entry ds.Entry) (ds.Function, error)
	RemoveFunction(ctx context.Context, fn ds.Function) error
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
	function, err := r.builder.Build(ctx, entry)
	if err != nil {
		return "", errors.Wrap(err, "Register Build error")
	}

	id, err = r.repo.SaveFunction(ctx, function)
	if err != nil {
		return "", errors.Wrap(err, "Save function error")
	}

	return id, nil
}

func (r *Registry) Get(ctx context.Context, id string) (ds.Function, error) {
	return r.repo.GetFunction(ctx, id)
}

// can corrupt data, fine for now
func (r *Registry) Delete(ctx context.Context, id string) error {

	fn, err := r.repo.GetFunction(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "err while getting function with id %v", id)
	}

	if err := r.builder.RemoveFunction(ctx, fn); err != nil {
		return errors.Wrapf(err, "err while removing artefacts for function id %v", id)
	}

	if err := r.repo.DeleteFunction(ctx, id); err != nil {
		return errors.Wrapf(err, "err while removing function entry with id %v from db", id)
	}

	return nil
}
