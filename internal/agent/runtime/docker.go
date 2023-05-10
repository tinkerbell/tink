package runtime

import (
	"context"
	"fmt"
	"io"
	"regexp"

	retry "github.com/avast/retry-go"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/go-logr/logr"
	"github.com/tinkerbell/tink/internal/agent"
	"github.com/tinkerbell/tink/internal/agent/runtime/internal"
	"github.com/tinkerbell/tink/internal/agent/workflow"
	"github.com/tinkerbell/tink/internal/ptr"
	"k8s.io/apimachinery/pkg/util/rand"
)

var _ agent.ContainerRuntime = &Docker{}

// Docker is a docker runtime that satisfies agent.ContainerRuntime.
type Docker struct {
	log    logr.Logger
	client *client.Client
}

// Run satisfies agent.ContainerRuntime.
func (d *Docker) Run(ctx context.Context, a workflow.Action) error {
	pullImage := func() error {
		// We need the image to be available before we can create a container.
		image, err := d.client.ImagePull(ctx, a.Image, types.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("docker: %w", err)
		}
		defer image.Close()

		// Docker requires everything to be read from the images ReadCloser for the image to actually
		// be pulled. We may want to log image pulls in a circular buffer somewhere for debugability.
		if _, err = io.Copy(io.Discard, image); err != nil {
			return fmt.Errorf("docker: %w", err)
		}

		return nil
	}

	err := retry.Do(pullImage, retry.Attempts(5), retry.DelayType(retry.BackOffDelay))
	if err != nil {
		return err
	}

	// TODO: Support all the other things on the action such as volumes.
	cfg := container.Config{
		Image: a.Image,
		Env:   toDockerEnv(a.Env),
	}

	failureFiles, err := internal.NewFailureFiles()
	if err != nil {
		return fmt.Errorf("create action failure files: %w", err)
	}
	defer failureFiles.Close()

	hostCfg := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: failureFiles.ReasonPath(),
				Target: ReasonMountPath,
			},
			{
				Type:   mount.TypeBind,
				Source: failureFiles.MessagePath(),
				Target: MessageMountPath,
			},
		},
	}

	containerName := toContainerName(a.ID)

	// Docker uses the entrypoint as the default command. The Tink Action Cmd property is modeled
	// as being the command launched in the container hence it is used as the entrypoint. Args
	// on the action are therefore the command portion in Docker.
	if a.Cmd != "" {
		cfg.Entrypoint = append(cfg.Entrypoint, a.Cmd)
	}
	if len(a.Args) > 0 {
		cfg.Cmd = append(cfg.Cmd, a.Args...)
	}

	// TODO: Figure out container logging. We probably want to save it somewhere for debugability.

	create, err := d.client.ContainerCreate(ctx, &cfg, &hostCfg, nil, nil, containerName)
	if err != nil {
		return fmt.Errorf("docker: %w", err)
	}

	// Always try to remove the container on exit.
	defer func() {
		// Force remove containers in an attempt to preserve space in memory constraints environments.
		// In rare cases this may create orphaned volumes that the Docker CLI won't clean up.
		opts := types.ContainerRemoveOptions{Force: true}

		// We can't use the context passed to Run() as it may have been cancelled so we use Background()
		// instead.
		err := d.client.ContainerRemove(context.Background(), create.ID, opts)
		if err != nil {
			d.log.Info("Couldn't remove container", "container_name", containerName, "error", err)
		}
	}()

	// Issue the wait with a 'next-exit' condition so we can await a response originating from
	// ContainerStart().
	waitBody, waitErr := d.client.ContainerWait(ctx, create.ID, container.WaitConditionNextExit)

	if err := d.client.ContainerStart(ctx, create.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("docker: %w", err)
	}

	select {
	case result := <-waitBody:
		if result.StatusCode == 0 {
			return nil
		}
		return failureFiles.ToError()

	case err := <-waitErr:
		return fmt.Errorf("docker: %w", err)

	case <-ctx.Done():
		// We can't use the context passed to Run() as its been cancelled.
		err := d.client.ContainerStop(context.Background(), create.ID, container.StopOptions{
			Timeout: ptr.Int(5),
		})
		if err != nil {
			d.log.Info("Failed to gracefully stop container", "error", err)
		}
		return fmt.Errorf("docker: %w", ctx.Err())
	}
}

var validContainerName = regexp.MustCompile(`[^a-zA-Z0-9_.-]`)

// toContainerName converts an action ID into a usable container name.
func toContainerName(actionID string) string {
	// Prepend 'tinkerbell_' so we guarantee the additional constraints on the first character.
	return fmt.Sprintf(
		"tinkerbell_%s_%s",
		validContainerName.ReplaceAllString(actionID, "_"),
		rand.String(6),
	)
}

func toDockerEnv(env map[string]string) []string {
	var de []string
	for k, v := range env {
		de = append(de, fmt.Sprintf("%v=%v", k, v))
	}
	return de
}

// NewDocker creates a new Docker instance.
func NewDocker(opts ...DockerOption) (*Docker, error) {
	o := &Docker{
		log: logr.Discard(),
	}

	var err error
	o.client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	for _, fn := range opts {
		fn(o)
	}

	return o, nil
}

// DockerOption defines optional configuration for a Docker instance.
type DockerOption func(*Docker)

// WithLogger returns an option to configure the logger on a Docker instance.
func WithLogger(log logr.Logger) DockerOption {
	return func(o *Docker) {
		if log.GetSink() == nil {
			return
		}
		o.log = log
	}
}

// WithClient returns an option to configure a Docker client on a Docker instance.
func WithClient(clnt *client.Client) DockerOption {
	return func(o *Docker) {
		if clnt == nil {
			return
		}
		o.client = clnt
	}
}
