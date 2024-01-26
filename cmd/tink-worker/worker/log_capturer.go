package worker

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-logr/logr"
)

// DockerLogCapturer is a LogCapturer that can stream docker container logs to an io.Writer.
type DockerLogCapturer struct {
	dockerClient client.ContainerAPIClient
	logger       logr.Logger
	writer       io.Writer
}

// getLogger is a helper function to get logging out of a context, or use the default logger.
func (l *DockerLogCapturer) getLogger(ctx context.Context) logr.Logger {
	loggerIface := ctx.Value(loggingContextKey)
	if loggerIface == nil {
		return l.logger
	}
	lg, _ := loggerIface.(logr.Logger)
	return lg
}

// NewDockerLogCapturer returns a LogCapturer that can stream container logs to a given writer.
func NewDockerLogCapturer(cli client.ContainerAPIClient, logger logr.Logger, writer io.Writer) *DockerLogCapturer {
	return &DockerLogCapturer{
		dockerClient: cli,
		logger:       logger,
		writer:       writer,
	}
}

// CaptureLogs streams container logs to the capturer's writer.
func (l *DockerLogCapturer) CaptureLogs(ctx context.Context, id string) {
	reader, err := l.dockerClient.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	})
	if err != nil {
		l.getLogger(ctx).Error(err, "failed to capture logs for container ", "containerID", id)
		return
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Fprintln(l.writer, scanner.Text())
	}
}
