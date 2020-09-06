package internal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
)

// RegistryConnDetails are the connection details for accessing a Docker
// registry and logging activities
type RegistryConnDetails struct {
	registry,
	user,
	pwd string
	logger log.Logger
}

// NewRegistryConnDetails creates a new RegistryConnDetails
func NewRegistryConnDetails(registry, user, pwd string, logger log.Logger) *RegistryConnDetails {
	return &RegistryConnDetails{
		registry: registry,
		user:     user,
		pwd:      pwd,
		logger:   logger,
	}
}

// NewClient uses the RegistryConnDetails to create a new Docker Client
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

// pullImage outputs to stdout the contents of the requested image (relative to the registry)
func (r *RegistryConnDetails) pullImage(ctx context.Context, cli *client.Client, image string) error {
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
	if _, err := io.Copy(os.Stdout, out); err != nil {
		return err
	}
	return nil
}
