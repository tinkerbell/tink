package worker

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
)

type fakeDockerLoggerClient struct {
	client.ContainerAPIClient
	content string
	err     error
}

func (c *fakeDockerLoggerClient) ContainerLogs(context.Context, string, types.ContainerLogsOptions) (io.ReadCloser, error) {
	if c.err != nil {
		return nil, c.err
	}
	return io.NopCloser(strings.NewReader(c.content)), nil
}

func newFakeDockerLoggerClient(content string, err error) *fakeDockerLoggerClient {
	return &fakeDockerLoggerClient{
		content: content,
		err:     err,
	}
}

func TestLogCapturer(t *testing.T) {
	cases := []struct {
		name    string
		writer  bytes.Buffer
		wanterr error
		content string
	}{
		{
			name:    "Content written to buffer",
			writer:  *bytes.NewBufferString(""),
			wanterr: nil,
			content: "Line1\nline2\n",
		},
		{
			name:    "empty buffer from error",
			writer:  *bytes.NewBufferString(""),
			wanterr: errors.New("Docker failure"),
			content: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logger := log.Test(t, "github.com/tinkerbell/tink")
			ctx := context.Background()
			clogger := NewDockerLogCapturer(
				newFakeDockerLoggerClient(tc.content, tc.wanterr),
				logger,
				&tc.writer)
			clogger.CaptureLogs(ctx, tc.name)
			got := tc.writer.String()
			if got != tc.content {
				t.Errorf("Wrong content written to buffer. Expected '%s', got '%s'", tc.content, got)
			}
		})
	}
}

func TestLogCapturerContextLogger(t *testing.T) {
	cases := []struct {
		name   string
		logger func() *log.Logger
		writer bytes.Buffer
	}{
		{
			name:   "no context logger",
			logger: nil,
		},
		{
			name: "with context logger",
			logger: func() *log.Logger {
				l := log.Test(t, "github.com/tinkerbell/tink/test")
				return &l
			},
			writer: *bytes.NewBufferString(""),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logger := log.Test(t, "github.com/tinkerbell/tink")
			ctx := context.Background()
			if tc.logger != nil {
				ctx = context.WithValue(ctx, loggingContextKey, tc.logger())
			}
			clogger := &DockerLogCapturer{
				newFakeDockerLoggerClient("", nil),
				logger,
				os.Stdout,
			}
			clogger.getLogger(ctx)
		})
	}
}
