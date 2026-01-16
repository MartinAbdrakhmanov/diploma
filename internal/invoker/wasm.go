package invoker

import (
	"bytes"
	"context"
	"fmt"
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
	wasmBytes, _ := os.ReadFile(fn.WasmPath)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	if timeout == 0 {
		timeout = ds.DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := i.r.InstantiateWithConfig(
		ctx,
		wasmBytes,
		wazero.NewModuleConfig().
			WithStdin(bytes.NewReader(input)).
			WithStdout(stdout).
			WithStderr(stderr),
	)
	execLog.FinishedAt = time.Now()
	execLog.DurationMs = execLog.FinishedAt.Sub(execLog.StartedAt).Milliseconds()
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
