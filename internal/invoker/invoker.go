package invoker

import (
	"context"
	"time"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/containerd/containerd"
	"github.com/tetratelabs/wazero"
)

type Invoker struct {
	client *containerd.Client
	r      wazero.Runtime
}

func New(
	client *containerd.Client,
	r wazero.Runtime,
) *Invoker {
	return &Invoker{
		client: client,
		r:      r,
	}
}

func (i *Invoker) Invoke(
	ctx context.Context,
	fn ds.Function,
	input []byte,
	timeout time.Duration,
) ([]byte, []byte, error) {

	switch fn.Runtime {
	case ds.DockerRuntime:
		return i.invokeDocker(ctx, fn, input, timeout)
	case ds.WasmRuntime:
		return i.invokeWasm(ctx, fn, input, timeout)
	}

	return nil, nil, ds.ErrInvalidRuntime
}
