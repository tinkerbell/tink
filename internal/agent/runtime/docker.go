package runtime

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/go-logr/logr"
	"github.com/tinkerbell/tink/internal/agent"
	"github.com/tinkerbell/tink/internal/agent/runtime/internal"
	"github.com/tinkerbell/tink/internal/agent/workflow"
)

var _ agent.ContainerRuntime = &Docker{}

type Docker struct {
	log    logr.Logger
	client *client.Client
}

func NewDocker() (*Docker, error) {
	// TODO: Offer an option to inject the client.
	clnt, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &Docker{
		// TODO: Fix the logger
		log:    logr.Discard(),
		client: clnt,
	}, nil
}

// Run satisfies agent.ContainerRuntime.
func (d *Docker) Run(ctx context.Context, a workflow.Action) error {
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
	// Close should be idempotent and we don't need the handle beyond this point.
	image.Close()

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

	containerName := toValidContainerName(a.ID)

	// Docker uses the entrypoint as the default command. The Tink Action Cmd property is modeled
	// as being the command launched in the container hence it is used as the entrypoint. Args
	// on the action are therefore the command portion in Docker.
	if a.Cmd != "" {
		cfg.Entrypoint = append(cfg.Entrypoint, a.Cmd)
	}
	if len(a.Args) > 0 {
		cfg.Cmd = append(cfg.Cmd, a.Args...)
	}

	// TODO: Figure out container logging.

	create, err := d.client.ContainerCreate(ctx, &cfg, &hostCfg, nil, nil, containerName)
	if err != nil {
		return fmt.Errorf("docker: %w", err)
	}

	// Always try to remove the container on exit.
	defer func() {
		err := d.client.ContainerRemove(ctx, create.ID, types.ContainerRemoveOptions{})
		if err != nil {
			d.log.Info("Couldn't remove container", "container_name", containerName, "error", err)
		}
	}()

	return dockerContainer{
		Log:    d.log.WithValues("container_id", create.ID),
		Client: d.client,
		ID:     create.ID,
		Files:  failureFiles,
	}.Run(ctx)
}

func toDockerEnv(env map[string]string) []string {
	var de []string
	for k, v := range env {
		de = append(de, fmt.Sprintf("%v=%v", k, v))
	}
	return de
}

var validContainerNameRegex = regexp.MustCompile(`[^a-zA-Z0-9_.-]`)

// toValidContainerName returns a valid container name for docker.
// Container names must satisfy [a-zA-Z0-9][a-zA-Z0-9_.-].
func toValidContainerName(name string) string {
	// Prepend 'action_' so we guarantee the additional constraints on the first character.
	return "action_" + validContainerNameRegex.ReplaceAllString(name, "_")
}

type dockerContainer struct {
	Log logr.Logger

	// Client used to run and shutdown the container.
	Client *client.Client

	// ID of the container. The container should have already been created hence this ID should
	// be available.
	ID string

	// Files are the failure files configured in the container and used to populate
	Files *internal.FailureFiles

	// StopTimeout is the time to wait for an action to exit when cancelled. When nil, the Docker
	// default timeout is used (10s).
	StopTimeout *time.Duration
}

func (c dockerContainer) Run(ctx context.Context) error {
	// Issue the wait with a 'next-exit' condition so we can await a response originating from
	// ContainerStart().
	waitBody, waitErr := c.Client.ContainerWait(ctx, c.ID, container.WaitConditionNextExit)

	if err := c.Client.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("docker: %w", err)
	}

	select {
	case result := <-waitBody:
		if result.StatusCode == 0 {
			return nil
		}
		return c.Files.ToError()

	case err := <-waitErr:
		return fmt.Errorf("docker: %w", err)

	case <-ctx.Done():
		var stopTimeout *time.Duration
		if c.StopTimeout != nil {
			stopTimeout = c.StopTimeout
		}
		if err := c.Client.ContainerStop(context.Background(), c.ID, stopTimeout); err != nil {
			c.Log.Info("Failed to gracefully stop container", "error", err)
		}
		return fmt.Errorf("docker: %w", ctx.Err())
	}
}
