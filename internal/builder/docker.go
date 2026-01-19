package builder

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-faster/errors"
)

// TODO change to local registry
func (b *Builder) buildDocker(ctx context.Context, name string, files map[string][]byte) (string, error) {
	// h := sha256.New()
	// for fname, content := range files {
	// 	io.WriteString(h, fname)
	// 	h.Write(content)
	// }
	// tag := hex.EncodeToString(h.Sum(nil))[:12]
	tag := hex.EncodeToString([]byte(time.Now().String()))[:12]
	image := fmt.Sprintf("mini-faas/%s:%s", name, tag)
	dir, err := prepareBuildContext(files)
	if err != nil {
		return "", err
	}

	defer os.RemoveAll(dir)

	if err := dockerBuild(ctx, dir, image); err != nil {
		return "", fmt.Errorf("docker build failed: %w", err)
	}

	if err := importToContainerd(image); err != nil {
		return "", fmt.Errorf("push to containerd failed: %w", err)
	}

	os.Remove(fmt.Sprintf("/tmp/%s.tar", strings.ReplaceAll(strings.ReplaceAll(image, "/", "_"), ":", "_")))

	image = "docker.io/" + image

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
	cmd := exec.CommandContext(ctx,
		"docker", "build",
		"-t", image,
		dir,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func importToContainerd(image string) error {
	safeName := strings.ReplaceAll(image, "/", "_")
	safeName = strings.ReplaceAll(safeName, ":", "_")

	tarPath := fmt.Sprintf("/tmp/%s.tar", safeName)

	cmd := exec.Command(
		"docker", "save",
		"-o", tarPath,
		image,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "docker save failed")
	}

	cmd = exec.Command(
		"sudo", "ctr", "images", "import", tarPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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
