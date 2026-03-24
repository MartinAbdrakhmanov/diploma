package cleaner

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
)

func (c *Cleaner) RemoveFunction(ctx context.Context, fn ds.Function) error {
	switch fn.Runtime {
	case ds.WasmRuntime:
		if fn.WasmPath != "" {
			if err := os.Remove(fn.WasmPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove wasm file: %w", err)
			}
		}
	case ds.DockerRuntime:
		if fn.Image != "" {

			if err := c.removeFromRegistry(ctx, fn.Image); err != nil {
				log.Printf("Warning: could not delete from registry: %v", err)
			}
			cmd := exec.CommandContext(ctx, "docker", "rmi", fn.Image)
			_ = cmd.Run()
		}
	}

	return nil
}

func (c *Cleaner) removeFromRegistry(ctx context.Context, image string) error {
	repo, tag, err := parseImage(image)
	if err != nil {
		return err
	}

	registryURL := "http://" + c.cfg.DockerRegistryPath

	digest, err := c.getManifestDigest(ctx, registryURL, repo, tag)
	if err != nil {
		return fmt.Errorf("failed to get digest: %w", err)
	}

	if err := c.deleteManifest(ctx, registryURL, repo, digest); err != nil {
		return fmt.Errorf("failed to delete manifest: %w", err)
	}

	log.Printf("Successfully removed image [%s:%s] by digest [%s]", repo, tag, digest)
	return nil
}

func parseImage(image string) (repo, tag string, err error) {
	firstSlash := strings.Index(image, "/")
	if firstSlash == -1 {
		return "", "", fmt.Errorf("invalid image format: %s", image)
	}

	repoWithTag := image[firstSlash+1:]
	lastColon := strings.LastIndex(repoWithTag, ":")

	repo = repoWithTag
	tag = "latest"
	if lastColon != -1 {
		repo = repoWithTag[:lastColon]
		tag = repoWithTag[lastColon+1:]
	}
	return repo, tag, nil
}

func (c *Cleaner) getManifestDigest(ctx context.Context, registryURL, repo, tag string) (string, error) {
	manifestURL := fmt.Sprintf("%s/v2/%s/manifests/%s", registryURL, repo, tag)

	req, err := http.NewRequestWithContext(ctx, "HEAD", manifestURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", strings.Join([]string{
		"application/vnd.oci.image.index.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.oci.image.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.list.v2+json",
	}, ", "))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return "", fmt.Errorf("digest header missing")
	}

	return digest, nil
}

func (c *Cleaner) deleteManifest(ctx context.Context, registryURL, repo, digest string) error {
	deleteURL := fmt.Sprintf("%s/v2/%s/manifests/%s", registryURL, repo, digest)

	delReq, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		return err
	}

	delResp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		return err
	}
	defer delResp.Body.Close()

	if delResp.StatusCode != http.StatusAccepted && delResp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status %d", delResp.StatusCode)
	}

	return nil
}
