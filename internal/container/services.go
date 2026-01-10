package container

import (
	"context"

	apigateway "github.com/MartinAbdrakhmanov/diploma/internal/api_gateway"
	"github.com/MartinAbdrakhmanov/diploma/internal/builder"
	functionregistry "github.com/MartinAbdrakhmanov/diploma/internal/function_registry"
	"github.com/MartinAbdrakhmanov/diploma/internal/invoker"
	"github.com/containerd/containerd"
	"github.com/go-faster/errors"
)

const (
	containerdPath = "/run/containerd/containerd.sock"
	baseURL        = "localhost" //???
)

func (c *Container) GetBuilderSvc(ctx context.Context) (*builder.Builder, error) {
	if c.builderSvc != nil {
		return c.builderSvc, nil
	}

	builder := builder.New()

	c.builderSvc = builder

	return c.builderSvc, nil
}

func (c *Container) GetInvokerSvc(ctx context.Context) (*invoker.Invoker, error) {
	if c.invokerSvc != nil {
		return c.invokerSvc, nil
	}

	client, err := containerd.New(containerdPath)
	if err != nil {
		return nil, errors.Wrap(err, "GetInvokerSvc containerd.New err")
	}
	c.closers = append(c.closers, func() {
		client.Close()
	})

	invoker := invoker.New(client)
	c.invokerSvc = invoker

	return c.invokerSvc, nil
}

func (c *Container) GetFunctionRegistry(ctx context.Context) (*functionregistry.Registry, error) {
	if c.functionRegistrySvc != nil {
		return c.functionRegistrySvc, nil
	}

	repo, err := c.GetRepo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetFunctionRegistry GetRepo err")
	}

	builder, err := c.GetBuilderSvc(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetFunctionRegistry GetBuilderSvc err")
	}

	functionRegistry := functionregistry.New(repo, builder)
	c.functionRegistrySvc = functionRegistry

	return c.functionRegistrySvc, nil
}

func (c *Container) GetApiGateway(ctx context.Context) (*apigateway.Gateway, error) {
	if c.apiGW != nil {
		return c.apiGW, nil
	}

	functionRegistry, err := c.GetFunctionRegistry(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetApiGateway GetFunctionRegistry err")
	}

	invoker, err := c.GetInvokerSvc(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetApiGateway GetInvokerSvc err")
	}

	apiGW := apigateway.New(functionRegistry, invoker, baseURL)
	c.apiGW = apiGW

	return c.apiGW, nil
}
