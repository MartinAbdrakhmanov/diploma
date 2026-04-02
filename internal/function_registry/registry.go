package functionregistry

import (
	"context"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/go-faster/errors"
)

type repository interface {
	SaveFunction(ctx context.Context, function ds.Function) (id string, err error)
	FunctionInfo(ctx context.Context, userID, id string) (function *ds.Function, err error)
	DeleteFunction(ctx context.Context, userID, id string) error
	FunctionStats(ctx context.Context, userID, functionID string) (ds.FunctionStats, error)
	UserFunctions(ctx context.Context, userID string) ([]ds.Function, error)
}

type builder interface {
	Build(ctx context.Context, entry ds.Entry) (ds.Function, error)
}

type cleaner interface {
	RemoveFunction(ctx context.Context, fn ds.Function) error
}

type Registry struct {
	repo    repository
	builder builder
	cleaner cleaner
}

func New(repo repository, builder builder, cleaner cleaner) *Registry {
	return &Registry{
		repo:    repo,
		builder: builder,
		cleaner: cleaner,
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

func (r *Registry) Get(ctx context.Context, userID, id string) (*ds.Function, error) {
	return r.repo.FunctionInfo(ctx, userID, id)
}

// can corrupt data, fine for now
func (r *Registry) Delete(ctx context.Context, userID, id string) error {

	fn, err := r.repo.FunctionInfo(ctx, userID, id)
	if err != nil {
		return errors.Wrapf(err, "err while getting function with id %v", id)
	}

	if fn == nil {
		return nil
	}

	if err := r.cleaner.RemoveFunction(ctx, *fn); err != nil {
		return errors.Wrapf(err, "err while removing artefacts for function id %v", id)
	}

	if err := r.repo.DeleteFunction(ctx, userID, id); err != nil {
		return errors.Wrapf(err, "err while removing function entry with id %v from db", id)
	}

	return nil
}

func (r *Registry) FunctionStats(ctx context.Context, userID, id string) (ds.FunctionStats, error) {
	stats, err := r.repo.FunctionStats(ctx, userID, id)
	if err != nil {
		return ds.FunctionStats{}, errors.Wrapf(err, "err while fetching stats for function %v", id)
	}
	return stats, nil
}

func (r *Registry) List(ctx context.Context, userID string) ([]ds.Function, error) {
	funcInfo, err := r.repo.UserFunctions(ctx, userID)
	if err != nil {
		return nil, errors.Wrapf(err, "err while fetching functions for user %v", userID)
	}
	return funcInfo, nil
}
