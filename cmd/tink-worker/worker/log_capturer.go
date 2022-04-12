package worker

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-logr/logr"
)

// DockerLogCapturer is a LogCapturer that can stream docker container logs to an io.Writer.
type DockerLogCapturer struct {
	dockerClient client.ContainerAPIClient
	logger       logr.Logger
	writer       io.Writer
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
	reader, err := l.dockerClient.ContainerLogs(ctx, id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	})
	if err != nil {
		l.logger.Error(err, "failed to capture logs for container ", "id", id)
		return
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Fprintln(l.writer, scanner.Text())
	}
}
