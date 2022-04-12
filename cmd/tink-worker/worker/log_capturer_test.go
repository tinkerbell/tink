package worker

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-logr/logr"
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
			logger := logr.Discard()
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
