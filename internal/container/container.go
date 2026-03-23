package container

import (
	apigateway "github.com/MartinAbdrakhmanov/diploma/internal/api_gateway"
	"github.com/MartinAbdrakhmanov/diploma/internal/builder"
	functionregistry "github.com/MartinAbdrakhmanov/diploma/internal/function_registry"
	"github.com/MartinAbdrakhmanov/diploma/internal/invoker"
	"github.com/MartinAbdrakhmanov/diploma/internal/repository"
	"github.com/MartinAbdrakhmanov/diploma/pkg/storage"
)

type Container struct {
	dbManager *storage.Client
	closers   []func()
	cfg       *config

	repo *repository.Repository

	builderSvc          *builder.Builder
	invokerSvc          *invoker.Invoker
	functionRegistrySvc *functionregistry.Registry
	apiGW               *apigateway.Gateway
}

func New(closers []func()) (*Container, error) {
	cfg, err := NewConfig()
	if err != nil {
		return nil, err
	}
	return &Container{
		closers: closers,
		cfg:     cfg,
	}, nil
}

func (c *Container) Close() {
	for i := len(c.closers) - 1; i >= 0; i-- {
		c.closers[i]()
	}
}
