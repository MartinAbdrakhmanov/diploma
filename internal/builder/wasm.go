package builder

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const wasmStoreDir = "/var/lib/mini-faas/wasm"

func (b *Builder) buildWasm(ctx context.Context, name string, files map[string][]byte) (string, error) {

	dir, err := prepareBuildContext(files)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	cmd := exec.CommandContext(ctx,
		"go", "build",
		"-o", "handler.wasm",
	)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GOOS=wasip1",
		"GOARCH=wasm",
	)

	if err := cmd.Run(); err != nil {
		return "", err
	}

	wasmPath := filepath.Join(dir, "handler.wasm")

	return storeWasm(name, wasmPath)
}

func storeWasm(functionID string, wasmPath string) (string, error) {
	if err := os.MkdirAll(wasmStoreDir, 0755); err != nil {
		return "", fmt.Errorf("create wasm store dir: %w", err)
	}

	dstPath := filepath.Join(wasmStoreDir, functionID+".wasm")

	src, err := os.Open(wasmPath)
	if err != nil {
		return "", fmt.Errorf("open source wasm: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("create destination wasm: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("copy wasm file: %w", err)
	}

	if err := os.Chmod(dstPath, 0444); err != nil {
		return "", fmt.Errorf("chmod wasm file: %w", err)
	}

	return dstPath, nil
}
