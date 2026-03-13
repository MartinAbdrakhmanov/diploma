package invoker

import (
	"bytes"
	"context"
	"os"
	"time"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/tetratelabs/wazero"
)

const pageSize = 64 * 1024

func (i *Invoker) invokeWasm(
	ctx context.Context,
	fn ds.Function,
	input []byte,
	timeout time.Duration,
) ([]byte, []byte, error, *ds.ExecLog) {

	execLog := &ds.ExecLog{
		FunctionID: fn.ID,
		StartedAt:  time.Now(),
		Status:     ds.StatusSuccess,
	}
	compiled, err := i.GetOrCompile(ctx, fn)
	if err != nil {
		return nil, nil, err, nil
	}
	defer compiled.Close(context.Background())
	execLog.InitTimeMs = time.Since(execLog.StartedAt).Milliseconds()

	if timeout == 0 {
		timeout = ds.DefaultTimeout
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	modCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	mod, err := i.r.InstantiateModule(modCtx, compiled, wazero.NewModuleConfig().
		WithStdin(bytes.NewReader(input)).
		WithStdout(stdout).
		WithStderr(stderr))

	execLog.FinishedAt = time.Now()
	execLog.DurationMs = time.Since(execLog.StartedAt).Milliseconds()
	execLog.ExecTimeMs = time.Since(execLog.StartedAt).Milliseconds() - execLog.InitTimeMs

	if err != nil {
		if modCtx.Err() == context.DeadlineExceeded {
			execLog.Status = ds.StatusTimeout
			execLog.ErrorMessage = ds.ErrFunctionTimeout.Error()
			return stdout.Bytes(), stderr.Bytes(),
				ds.ErrFunctionTimeout, execLog
		}

		execLog.Status = ds.StatusError
		execLog.ErrorMessage = err.Error()
		return stdout.Bytes(), stderr.Bytes(), err, execLog
	}
	if mem := mod.Memory(); mem != nil {
		currentPages, _ := mem.Grow(0)
		execLog.MaxMemoryBytes = uint64(currentPages) * uint64(pageSize)
	}

	defer mod.Close(context.Background())

	return stdout.Bytes(), stderr.Bytes(), nil, execLog
}

func (i *Invoker) GetOrCompile(ctx context.Context, fn ds.Function) (wazero.CompiledModule, error) {
	wasmBytes, err := os.ReadFile(fn.WasmPath)
	if err != nil {
		return nil, err
	}
	compiled, err := i.r.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, err
	}

	return compiled, nil
}
