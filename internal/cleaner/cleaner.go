package cleaner

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
)

const (
	oldForCache = 7 * 24 * time.Hour
)

type repo interface {
	FunctionLastCalledAt(ctx context.Context, functionID string) (time.Time, error)
	DeleteFunction(ctx context.Context, userID, id string) error
	GetExpiredFunctions(ctx context.Context, retentionInterval time.Duration) ([]ds.Function, error)
}

type Config struct {
	DockerRegistryPath string
	WasmCacheDir       string
}

type Cleaner struct {
	repo              repo
	cfg               Config
	cleanupInterval   time.Duration
	retentionInterval time.Duration
}

func New(
	repo repo,
	cfg Config,
	cleanupInterval time.Duration,
	retentionInterval time.Duration,
) *Cleaner {
	return &Cleaner{
		repo:              repo,
		cfg:               cfg,
		cleanupInterval:   cleanupInterval,
		retentionInterval: retentionInterval,
	}
}

func (c *Cleaner) Run(ctx context.Context) {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.performCleanup(ctx)
		}
	}
}

func (c *Cleaner) performCleanup(ctx context.Context) {
	log.Println("Starting scheduled cleanup...")

	c.cleanupOldFunctions(ctx)

	c.runRegistryGC(ctx)

	c.cleanupWasmCache()
}

func (c *Cleaner) cleanupWasmCache() {
	err := filepath.Walk(c.cfg.WasmCacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if time.Since(info.ModTime()) > oldForCache {
			os.Remove(path)
		}
		return nil
	})
	if err != nil {
		log.Printf("Wasm cache cleanup error: %v", err)
	}
}

func (c *Cleaner) cleanupOldFunctions(ctx context.Context) {
	oldFunctions, err := c.repo.GetExpiredFunctions(ctx, c.retentionInterval)
	if err != nil {
		log.Printf("error while gettion old functions info %v", err)
	}

	for _, fn := range oldFunctions {
		log.Printf("Removing expired function: %s", fn.Name)
		c.RemoveFunction(ctx, fn)
		c.repo.DeleteFunction(ctx, fn.UserId, fn.ID)
	}
}

func (c *Cleaner) runRegistryGC(ctx context.Context) {
	cmd := exec.CommandContext(ctx, "docker", "exec", "registry",
		"bin/registry", "garbage-collect", "/etc/docker/registry/config.yml")

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Registry GC failed: %v, output: %s", err, string(out))
	} else {
		log.Println("Registry GC completed successfully")
	}
}
