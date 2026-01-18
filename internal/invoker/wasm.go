package invoker

import (
	"bytes"
	"context"
	"fmt"
	"log"
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
	compiled, err := i.wasmCache.GetOrCompile(ctx, i.r, fn)
	if err != nil {
		return nil, nil, err, nil
	}
	log.Printf("GetOrCompile duration %v", time.Since(getCompTime).Milliseconds())

	if timeout == 0 {
		timeout = ds.DefaultTimeout
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// _, err := i.r.InstantiateWithConfig(
	// 	ctx,
	// 	wasmBytes,
	// 	wazero.NewModuleConfig().
	// 		WithStdin(bytes.NewReader(input)).
	// 		WithStdout(stdout).
	// 		WithStderr(stderr),
	// )

	initTime := time.Now()
	_, err = i.r.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().
		WithStdin(bytes.NewReader(input)).
		WithStdout(stdout).
		WithStderr(stderr))

	execLog.FinishedAt = time.Now()
	execLog.DurationMs = time.Since(execLog.StartedAt).Milliseconds()

	log.Printf("InstantiateModule duration %v", time.Since(initTime).Milliseconds())
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			execLog.Status = ds.StatusTimeout
			execLog.ErrorMessage = "function timeout" // TO DO fix
			return stdout.Bytes(), stderr.Bytes(),
				fmt.Errorf("function timeout"), execLog
		}

		execLog.Status = ds.StatusError
		execLog.ErrorMessage = err.Error()
		return stdout.Bytes(), stderr.Bytes(), err, execLog
	}

	return stdout.Bytes(), stderr.Bytes(), nil, execLog
}
