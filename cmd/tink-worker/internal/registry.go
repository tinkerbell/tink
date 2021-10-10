package internal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
)

// RegistryConnDetails are the connection details for accessing a Docker
// registry and logging activities.
type RegistryConnDetails struct {
	registry,
	user,
	pwd string
	logger log.Logger
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

// NewRegistryConnDetails creates a new RegistryConnDetails.
func NewRegistryConnDetails(registry, user, pwd string, logger log.Logger) *RegistryConnDetails {
	return &RegistryConnDetails{
		registry: registry,
		user:     user,
		pwd:      pwd,
		logger:   logger,
	}
}

// NewClient uses the RegistryConnDetails to create a new Docker Client.
func (r *RegistryConnDetails) NewClient() (*client.Client, error) {
	if r.registry == "" {
		return nil, errors.New("required DOCKER_REGISTRY")
	}
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Wrap(err, "DOCKER CLIENT")
	}

	return c, nil
}

type imagePuller interface {
	ImagePull(context.Context, string, types.ImagePullOptions) (io.ReadCloser, error)
}

// pullImage outputs to stdout the contents of the requested image (relative to the registry).
func (r *RegistryConnDetails) pullImage(ctx context.Context, cli imagePuller, image string) error {
	authConfig := types.AuthConfig{
		Username:      r.user,
		Password:      r.pwd,
		ServerAddress: r.registry,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return errors.Wrap(err, "DOCKER AUTH")
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	out, err := cli.ImagePull(ctx, r.registry+"/"+image, types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		return errors.Wrap(err, "DOCKER PULL")
	}
	defer out.Close()
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
