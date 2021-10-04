package internal

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/packethost/pkg/log"
	"github.com/stretchr/testify/assert"
)

func setupTestLogger(t *testing.T) log.Logger {
	t.Helper()

	service := "github.com/tinkerbell/tink"
	logger, err := log.Init(service)
	if err != nil {
		t.Fatal(err)
	}
	return logger
}

type imagePullerMock struct {
	stringReadCloser io.ReadCloser
	imagePullErr     error
}

func (d *imagePullerMock) ImagePull(_ context.Context, _ string, _ types.ImagePullOptions) (io.ReadCloser, error) {
	return d.stringReadCloser, d.imagePullErr
}

func TestPullImageAnyFailure(t *testing.T) {
	for _, test := range []struct {
		testName         string
		testString       string
		testImagePullErr error
		testErr          error
	}{
		{
			testName:         "success",
			testString:       "{\"status\": \"hello\",\"error\":\"\"}{\"status\":\"world\",\"error\":\"\"}",
			testImagePullErr: nil,
			testErr:          nil,
		},
		{
			testName:         "fail",
			testString:       "{\"error\": \"\"}",
			testImagePullErr: errors.New("Tested, failure of the image pull"),
			testErr:          errors.New("DOCKER PULL: Tested, failure of the image pull"),
		},
		{
			testName:         "fail_partial",
			testString:       "{\"status\": \"hello\",\"error\":\"\"}{\"status\":\"world\",\"error\":\"Tested, failure of No space left on device\"}",
			testImagePullErr: nil,
			testErr:          errors.New("DOCKER PULL: Tested, failure of No space left on device"),
		},
	} {
		t.Run(test.testName, func(t *testing.T) {
			ctx := context.Background()
			rcon := NewRegistryConnDetails("test", "testUser", "testPwd", setupTestLogger(t))
			stringReader := strings.NewReader(test.testString)
			cli := &imagePullerMock{
				stringReadCloser: ioutil.NopCloser(stringReader),
				imagePullErr:     test.testImagePullErr,
			}
			err := rcon.pullImage(ctx, cli, test.testName)
			if test.testErr != nil {
				assert.Equal(t, err.Error(), test.testErr.Error())
			} else {
				assert.Equal(t, err, test.testErr)
			}
		})
	}
}
