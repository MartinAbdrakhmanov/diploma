package builder

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
)

func (b *Builder) RemoveFunction(ctx context.Context, fn ds.Function) error {
	switch fn.Runtime {
	case ds.WasmRuntime:
		if fn.WasmPath != "" {
			if err := os.Remove(fn.WasmPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove wasm file: %w", err)
			}
		}
	case ds.DockerRuntime:
		if fn.Image != "" {

			// if err := b.removeFromRegistry(ctx, fn.Image); err != nil {
			// 	log.Printf("Warning: could not delete from registry: %v", err)
			// }
			// TODO изменить удаление из registry, сейчас ультра костыльно удаляется папка
			if err := b.removeRepositoryFromDisk(fn.Image); err != nil {
				log.Printf("Warning: could not delete from registry: %v", err)
			}
			cmd := exec.CommandContext(ctx, "docker", "rmi", fn.Image)
			_ = cmd.Run()
		}
	}

	return nil
}

func (b *Builder) removeRepositoryFromDisk(image string) error {
	firstSlash := strings.Index(image, "/")
	if firstSlash == -1 {
		return fmt.Errorf("invalid image format: %s", image)
	}
	rest := image[firstSlash+1:] // mini-faas/echotestdelete:617b33454569

	lastColon := strings.LastIndex(rest, ":")
	var repoPath string
	if lastColon == -1 {
		repoPath = rest
	} else {
		repoPath = rest[:lastColon] // mini-faas/echotestdelete
	}

	baseRegistryPath := "./local/registry-data/docker/registry/v2/repositories"

	fullPath := filepath.Join(baseRegistryPath, repoPath)

	log.Printf("Registry Hard Cleanup: Removing repository folder: %s", fullPath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Printf("Repository folder does not exist, skipping: %s", fullPath)
		return nil
	}
	err := os.RemoveAll(fullPath)
	if err != nil {
		return fmt.Errorf("failed to physically remove repository: %w", err)
	}

	log.Printf("Successfully wiped [%s] from registry disk", repoPath)
	return nil
}
