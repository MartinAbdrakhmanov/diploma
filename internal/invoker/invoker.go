package invoker

import (
	"context"
	"log"
	"time"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/containerd/containerd"
	"github.com/tetratelabs/wazero"
)

type repository interface {
	SaveLog(ctx context.Context, log ds.ExecLog) error
}

type Invoker struct {
	client    *containerd.Client
	r         wazero.Runtime
	repo      repository
	wasmCache *WasmCache
}

func New(
	client *containerd.Client,
	r wazero.Runtime,
	repo repository,
) *Invoker {
	return &Invoker{
		client:    client,
		r:         r,
		repo:      repo,
		wasmCache: NewWasmCache(),
	}
}

func (i *Invoker) Invoke(
	ctx context.Context,
	fn ds.Function,
	input []byte,
	timeout time.Duration,
) (stdout []byte, stderr []byte, err error) {

	var execLog *ds.ExecLog

	switch fn.Runtime {
	case ds.DockerRuntime:
		stdout, stderr, err, execLog = i.invokeDocker(ctx, fn, input, timeout)
	case ds.WasmRuntime:
		stdout, stderr, err, execLog = i.invokeWasm(ctx, fn, input, timeout)
	default:
		return nil, nil, ds.ErrInvalidRuntime
	}

	if err := i.repo.SaveLog(ctx, *execLog); err != nil {
		log.Printf("error while saving log for fn id %s, %v", execLog.FunctionID, err)
	}

	return stdout, stderr, err
}
