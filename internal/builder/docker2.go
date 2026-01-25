package builder

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func (b *Builder) buildDocker(ctx context.Context, name string, files map[string][]byte) (string, error) {
	// h := sha256.New()
	// for fname, content := range files {
	// 	h.Write([]byte(fname))
	// 	h.Write(content)
	// }
	// tag := hex.EncodeToString(h.Sum(nil))[:12]
	tag := hex.EncodeToString([]byte(time.Now().String()))[1:13]
	registry := "localhost:5000" //move to env
	image := fmt.Sprintf("%s/mini-faas/%s:%s", registry, name, tag)

	dir, err := prepareBuildContext(files)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	if err := dockerBuild(ctx, dir, image); err != nil {
		return "", fmt.Errorf("docker build failed: %w", err)
	}

	if err := pushToRegistry(ctx, image); err != nil {
		return "", fmt.Errorf("docker push failed: %w", err)
	}

	return image, nil
}

func prepareBuildContext(files map[string][]byte) (string, error) {
	dir, err := os.MkdirTemp("", "mini-faas-build-*")
	if err != nil {
		return "", err
	}

	for fname, content := range files {
		target := filepath.Join(dir, fname)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(target, content, 0644); err != nil {
			return "", err
		}
	}

	dockerfilePath := filepath.Join(dir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		if err := os.WriteFile(dockerfilePath, []byte(defaultGoDockerfile), 0644); err != nil {
			return "", err
		}
	}

	return dir, nil
}

func dockerBuild(ctx context.Context, dir, image string) error {
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", image, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func pushToRegistry(ctx context.Context, image string) error {
	cmd := exec.CommandContext(ctx, "docker", "push", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker push failed: %w", err)
	}
	return nil
}

const defaultGoDockerfile = `FROM golang:1.25-alpine AS build
WORKDIR /app
COPY . .
RUN go mod download && go build -o handler main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/handler /handler
ENTRYPOINT ["/handler"]
`
