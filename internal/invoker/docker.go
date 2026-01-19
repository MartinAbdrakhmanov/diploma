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
) ([]byte, []byte, error, *ds.ExecLog) {
	execLog := &ds.ExecLog{
		FunctionID: fn.ID,
		StartedAt:  time.Now(),
		Status:     ds.StatusSuccess,
	}

	ctx = namespaces.WithNamespace(ctx, "default")

	image, err := i.client.GetImage(ctx, fn.Image)
	if err != nil {
		image, err = i.client.Pull(
			ctx,
			fn.Image,
			containerd.WithPullUnpack,
		)
		if err != nil {
			return nil, nil, err, nil
		}
	}

	snapshotID := fn.ID + "-snap-" + time.Now().Format("150405.000")

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
		return nil, nil, err, nil
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)

	var stdout, stderr bytes.Buffer
	stdin := bytes.NewReader(input)

	task, err := container.NewTask(
		ctx,
		cio.NewCreator(
			cio.WithStreams(stdin, &stdout, &stderr),
		),
	)
	if err != nil {
		return nil, nil, err, nil
	}
	defer task.Delete(ctx, containerd.WithProcessKill)

	if err := task.Start(ctx); err != nil {
		return nil, nil, err, nil
	}

	task.CloseIO(ctx, containerd.WithStdinCloser)

	waitC, err := task.Wait(ctx)
	if err != nil {
		return nil, nil, err, nil
	}

	if timeout == 0 {
		timeout = ds.DefaultTimeout
	}

	var (
		execErr error
	)
	select {
	case status := <-waitC:
		if status.ExitCode() != 0 {
			execLog.ExitCode = status.ExitCode()

			execErr = fmt.Errorf("non-zero exit code: %d", status.ExitCode())
			execLog.ErrorMessage = execErr.Error()

			execLog.Status = ds.StatusError
		}
	case <-time.After(timeout):
		task.Kill(ctx, syscall.SIGKILL)
		execErr = fmt.Errorf("function timeout")
		execLog.ErrorMessage = execErr.Error()
		execLog.Status = ds.StatusTimeout
	}
	execLog.FinishedAt = time.Now()
	execLog.DurationMs = time.Since(execLog.StartedAt).Milliseconds()

	return stdout.Bytes(), stderr.Bytes(), execErr, execLog
}
