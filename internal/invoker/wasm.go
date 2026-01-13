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
) ([]byte, []byte, error) {
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

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return stdout.Bytes(), stderr.Bytes(),
				fmt.Errorf("function timeout")
		}
		return stdout.Bytes(), stderr.Bytes(), err
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}
