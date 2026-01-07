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

// BuildFunction принимает map файлов и собирает Docker image.
// files: ключ — относительный путь (например "main.go", "go.mod"), значение — содержимое файла
func Build(ctx context.Context, name string, files map[string][]byte) (string, error) {
	// 1) Создаём детерминированный тег через sha256 содержимого
	// h := sha256.New()
	// for fname, content := range files {
	// 	io.WriteString(h, fname)
	// 	h.Write(content)
	// }
	// tag := hex.EncodeToString(h.Sum(nil))[:12]
	tag := hex.EncodeToString([]byte(time.Now().String()))[:12]
	image := fmt.Sprintf("mini-faas/%s:%s", name, tag)

	// Создаём временную папку build context
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

	return image, nil
}

func prepareBuildContext(files map[string][]byte) (string, error) {
	// 2) Создаём временную папку build context
	dir, err := os.MkdirTemp("", "mini-faas-build-*")
	if err != nil {
		return "", err
	}

	// 3) Записываем все файлы в build context
	for fname, content := range files {
		target := filepath.Join(dir, fname)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(target, content, 0644); err != nil {
			return "", err
		}
	}

	// 4) Проверяем, есть ли Dockerfile, если нет — пишем дефолтный
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
	cmd := exec.Command(
		"sh", "-c",
		fmt.Sprintf("docker save %s -o /tmp/%[1]s.tar", image),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command(
		"sh", "-c",
		fmt.Sprintf("sudo ctr images import /tmp/%s.tar", image),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Default Dockerfile для Go функций
const defaultGoDockerfile = `FROM golang:1.25-alpine AS build
WORKDIR /app
COPY . .
RUN go mod download && go build -o handler main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/handler /handler
ENTRYPOINT ["/handler"]
`
