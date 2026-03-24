package invoker

import (
	"context"
	"log"
	"time"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/MartinAbdrakhmanov/diploma/internal/metrics"
	"github.com/containerd/containerd"
	"github.com/tetratelabs/wazero"
)

type repository interface {
	SaveLog(ctx context.Context, log ds.ExecLog) error
	UpdateFunction(ctx context.Context, functionID string) error
}

type Invoker struct {
	client *containerd.Client
	r      wazero.Runtime
	repo   repository
}

func New(
	client *containerd.Client,
	r wazero.Runtime,
	repo repository,
) *Invoker {
	return &Invoker{
		client: client,
		r:      r,
		repo:   repo,
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

	if execLog != nil {
		if err := i.repo.SaveLog(ctx, *execLog); err != nil {
			log.Printf("error while saving log for fn id %s, %v", execLog.FunctionID, err)
		}
		if err := i.repo.UpdateFunction(ctx, execLog.FunctionID); err != nil {
			log.Printf("error while updating function %v: %v", execLog.FunctionID, err)
		}
		metrics.FunctionDurationObserve(fn, float64(execLog.DurationMs))
		metrics.FunctionInvocationInc(fn, execLog.Status)
	}

	return stdout, stderr, err
}
