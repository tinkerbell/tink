package worker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"path"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/pkg/errors"
)

// RegistryConnDetails are the connection details for accessing a Docker registry.
type RegistryConnDetails struct {
	Registry string
	Username string
	Password string
}

// ImagePullStatus is the status of the downloaded Image chunk.
type ImagePullStatus struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

// PullImage outputs to stdout the contents of the requested image (relative to the registry).
// If a pull fails but the image already exists then we will return a nil error.
func (m *containerManager) PullImage(ctx context.Context, img string) error {
	l := m.getLogger(ctx)
	authConfig := registry.AuthConfig{
		Username:      m.registryDetails.Username,
		Password:      m.registryDetails.Password,
		ServerAddress: m.registryDetails.Registry,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return errors.Wrap(err, "DOCKER AUTH")
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	out, err := m.cli.ImagePull(ctx, path.Join(m.registryDetails.Registry, img), image.PullOptions{RegistryAuth: authStr})
	if err != nil {
		if _, _, err := m.cli.ImageInspectWithRaw(ctx, path.Join(m.registryDetails.Registry, img)); err == nil {
			return nil
		}
		return errors.Wrap(err, "DOCKER PULL")
	}
	defer func() {
		if err := out.Close(); err != nil {
			l.Error(err, "")
		}
	}()
	fd := json.NewDecoder(out)
	var status *ImagePullStatus
	for {
		if err := fd.Decode(&status); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return errors.Wrap(err, "DOCKER PULL")
		}
		if status.Error != "" {
			return errors.Wrap(errors.New(status.Error), "DOCKER PULL")
		}
	}
	return nil
}
