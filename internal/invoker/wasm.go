package invoker

import (
	"bytes"
	"context"
	"log"
	"os"
	"time"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/tetratelabs/wazero"
)

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
	getCompTime := time.Now()
	compiled, err := i.GetOrCompile(ctx, fn)
	if err != nil {
		return nil, nil, err, nil
	}
	defer compiled.Close(context.Background())
	log.Printf("GetOrCompile duration %v", time.Since(getCompTime).Milliseconds())

	if timeout == 0 {
		timeout = ds.DefaultTimeout
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	modCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	initTime := time.Now()
	mod, err := i.r.InstantiateModule(modCtx, compiled, wazero.NewModuleConfig().
		WithStdin(bytes.NewReader(input)).
		WithStdout(stdout).
		WithStderr(stderr))

	execLog.FinishedAt = time.Now()
	execLog.DurationMs = time.Since(execLog.StartedAt).Milliseconds()

	log.Printf("InstantiateModule duration %v", time.Since(initTime).Milliseconds())
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
