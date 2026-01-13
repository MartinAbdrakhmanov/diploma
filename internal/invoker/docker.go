package invoker

import (
	"bytes"
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
)

func (i *Invoker) invokeDocker(
	ctx context.Context,
	fn ds.Function,
	input []byte,
	timeout time.Duration,
) ([]byte, []byte, error) {

	ctx = namespaces.WithNamespace(ctx, "default")

	// 1. Pull image (idempotent)
	image, err := i.client.GetImage(ctx, fn.Image)
	if err != nil {
		image, err = i.client.Pull(
			ctx,
			fn.Image,
			containerd.WithPullUnpack,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	// 2. Unique snapshot per invocation
	snapshotID := fn.ID + "-snap-" + time.Now().Format("150405.000")

	// 3. Create container
	container, err := i.client.NewContainer(
		ctx,
		fn.ID,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(snapshotID, image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			// oci.WithProcessArgs(fn.Args...),
		),
	)
	if err != nil {
		return nil, nil, err
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)

	// 4. IO
	var stdout, stderr bytes.Buffer
	stdin := bytes.NewReader(input)

	task, err := container.NewTask(
		ctx,
		cio.NewCreator(
			cio.WithStreams(stdin, &stdout, &stderr),
		),
	)
	if err != nil {
		return nil, nil, err
	}
	defer task.Delete(ctx, containerd.WithProcessKill)

	// 5. Start
	if err := task.Start(ctx); err != nil {
		return nil, nil, err
	}

	// 6. Close stdin (EOF)
	task.CloseIO(ctx, containerd.WithStdinCloser)

	// 7. Wait with timeout
	waitC, err := task.Wait(ctx)
	if err != nil {
		return nil, nil, err
	}

	if timeout == 0 {
		timeout = ds.DefaultTimeout
	}
	select {
	case status := <-waitC:
		if status.ExitCode() != 0 {
			return stdout.Bytes(), stderr.Bytes(),
				fmt.Errorf("non-zero exit code: %d", status.ExitCode())
		}
	case <-time.After(timeout):
		task.Kill(ctx, syscall.SIGKILL)
		return stdout.Bytes(), stderr.Bytes(),
			fmt.Errorf("function timeout")
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}
