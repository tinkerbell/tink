package cmd

import (
	"os"
	"strings"
	"time"

	dockercli "github.com/docker/docker/client"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-worker/worker"
	"github.com/tinkerbell/tink/protos/workflow"
)

const (
	defaultRetryIntervalSeconds = 3
	defaultRetryCount           = 3
	defaultMaxFileSize          = 10 * 1024 * 1024 // 10MB
	defaultTimeoutMinutes       = 60
)

// NewRootCommand creates a new Tink Worker Cobra root command.
func NewRootCommand(version string, logger log.Logger) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "tink-worker",
		Short:   "Tink Worker",
		Version: version,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initViper(logger, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			retryInterval := viper.GetDuration("retry-interval")
			retries := viper.GetInt("max-retry")
			workerID := viper.GetString("id")
			maxFileSize := viper.GetInt64("max-file-size")
			user := viper.GetString("registry-username")
			pwd := viper.GetString("registry-password")
			registry := viper.GetString("docker-registry")
			useAbsoluteImageURI := viper.GetBool("use-absolute-image-uri")
			captureActionLogs := viper.GetBool("capture-action-logs")

			logger.With("version", version).Info("starting")

			conn, err := client.NewClientConn(
				viper.GetString("tinkerbell-grpc-authority"),
				viper.GetBool("tinkerbell-tls"),
			)
			if err != nil {
				return err
			}
			workflowClient := workflow.NewWorkflowServiceClient(conn)

			dockerClient, err := dockercli.NewClientWithOpts(dockercli.FromEnv, dockercli.WithAPIVersionNegotiation())
			if err != nil {
				return err
			}
			containerManager := worker.NewContainerManager(
				logger,
				dockerClient,
				worker.RegistryConnDetails{
					Registry:            registry,
					Username:            user,
					Password:            pwd,
					UseAbsoluteImageURI: useAbsoluteImageURI,
				})

			logCapturer := worker.NewDockerLogCapturer(dockerClient, logger, os.Stdout)

			w := worker.NewWorker(
				workerID,
				workflowClient,
				containerManager,
				logCapturer,
				logger,
				worker.WithMaxFileSize(maxFileSize),
				worker.WithRetries(retryInterval, retries),
				worker.WithLogCapture(captureActionLogs),
				worker.WithPrivileged(true))

			err = w.ProcessWorkflowActions(cmd.Context())
			if err != nil {
				return errors.Wrap(err, "worker Finished with error")
			}
			return nil
		},
	}

	rootCmd.Flags().Duration("retry-interval", defaultRetryIntervalSeconds*time.Second, "Retry interval in seconds (RETRY_INTERVAL)")
	rootCmd.Flags().Duration("timeout", defaultTimeoutMinutes*time.Minute, "Max duration to wait for worker to complete. Set to '0' for no timeout (TIMEOUT)")
	rootCmd.Flags().Int("max-retry", defaultRetryCount, "Maximum number of retries to attempt (MAX_RETRY)")
	rootCmd.Flags().Int64("max-file-size", defaultMaxFileSize, "Maximum file size in bytes (MAX_FILE_SIZE)")
	rootCmd.Flags().Bool("capture-action-logs", true, "Capture action container output as part of worker logs")
	rootCmd.Flags().Bool("tinkerbell-tls", true, "Connect to server via TLS or not (TINKERBELL_TLS)")
	rootCmd.Flags().StringP("docker-registry", "r", "", "Sets the Docker registry (DOCKER_REGISTRY)")
	rootCmd.Flags().StringP("registry-username", "u", "", "Sets the registry username (REGISTRY_USERNAME)")
	rootCmd.Flags().StringP("registry-password", "p", "", "Sets the registry-password (REGISTRY_PASSWORD)")
	rootCmd.Flags().BoolP("use-absolute-image-uri", "a", false, "Do not prepend docker_registry to template action images (USE_ABSOLUTE_IMAGE_URI)")

	must := func(err error) {
		if err != nil {
			logger.Fatal(err)
		}
	}

	rootCmd.Flags().StringP("id", "i", "", "Sets the worker id (ID)")
	must(rootCmd.MarkFlagRequired("id"))

	rootCmd.Flags().String("tinkerbell-grpc-authority", "", "tink server grpc endpoint (TINKERBELL_GRPC_AUTHORITY)")
	must(rootCmd.MarkFlagRequired("tinkerbell-grpc-authority"))

	_ = viper.BindPFlags(rootCmd.Flags())

	return rootCmd
}

// initViper initializes Viper  configured to read in configuration files
// (from various paths with content type specific filename extensions) and loads
// environment variables.
func initViper(logger log.Logger, cmd *cobra.Command) error {
	viper.AutomaticEnv()
	viper.SetConfigName("tink-worker")
	viper.AddConfigPath("/etc/tinkerbell")
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.With("configFile", viper.ConfigFileUsed()).Error(err, "could not load config file")
			return err
		}
		logger.Info("no config file found")
	} else {
		logger.With("configFile", viper.ConfigFileUsed()).Info("loaded config file")
	}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) {
			_ = cmd.Flags().SetAnnotation(f.Name, cobra.BashCompOneRequiredFlag, []string{"false"})
		}
	})

	return nil
}
