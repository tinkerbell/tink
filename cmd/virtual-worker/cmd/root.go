package cmd

import (
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	tinkWorker "github.com/tinkerbell/tink/cmd/tink-worker/worker"
	"github.com/tinkerbell/tink/cmd/virtual-worker/worker"
	"github.com/tinkerbell/tink/internal/client"
	"github.com/tinkerbell/tink/internal/proto"
	"go.uber.org/zap"
)

const (
	defaultRetryIntervalSeconds = 3
	defaultRetryCount           = 3
	defaultMaxFileSize          = 10 * 1024 * 1024 // 10MB
)

// NewRootCommand creates a new Virtual Worker Cobra root command.
func NewRootCommand(version string) *cobra.Command {
	zlog, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger := zapr.NewLogger(zlog).WithName("github.com/tinkerbell/tink")

	rootCmd := &cobra.Command{
		Use:   "virtual-worker",
		Short: "Virtual Tink Worker",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return createViper(logger, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			retryInterval := viper.GetDuration("retry-interval")
			retries := viper.GetInt("max-retry")
			workerID := viper.GetString("id")
			maxFileSize := viper.GetInt64("max-file-size")
			captureActionLogs := viper.GetBool("capture-action-logs")
			sleepMin := viper.GetDuration("sleep-min")
			sleepJitter := viper.GetDuration("sleep-jitter")

			logger.Info("starting", "version", version)

			conn, err := client.NewClientConn(
				viper.GetString("tinkerbell-grpc-authority"),
				viper.GetBool("tinkerbell-tls"),
			)
			if err != nil {
				return err
			}
			workflowClient := proto.NewWorkflowServiceClient(conn)

			containerManager := worker.NewFakeContainerManager(logger, sleepMin, sleepJitter)
			logCapturer := worker.NewEmptyLogCapturer()

			w := tinkWorker.NewWorker(
				workerID,
				workflowClient,
				containerManager,
				logCapturer,
				logger,
				tinkWorker.WithMaxFileSize(maxFileSize),
				tinkWorker.WithRetries(retryInterval, retries),
				tinkWorker.WithDataDir("./worker"),
				tinkWorker.WithLogCapture(captureActionLogs))

			err = w.ProcessWorkflowActions(cmd.Context())
			if err != nil {
				return errors.Wrap(err, "worker Finished with error")
			}
			return nil
		},
	}

	rootCmd.Flags().Duration("retry-interval", defaultRetryIntervalSeconds*time.Second, "Retry interval in seconds (RETRY_INTERVAL)")
	rootCmd.Flags().Int("max-retry", defaultRetryCount, "Maximum number of retries to attempt (MAX_RETRY)")
	rootCmd.Flags().Int64("max-file-size", defaultMaxFileSize, "Maximum file size in bytes (MAX_FILE_SIZE)")
	rootCmd.Flags().Bool("capture-action-logs", true, "Capture action container output as part of worker logs")
	rootCmd.Flags().Duration("sleep-min", time.Second*4, "The minimum amount of time to sleep during faked docker operations")
	rootCmd.Flags().Duration("sleep-jitter", time.Second*2, "The amount of jitter to add during faked docker operations")

	must := func(err error) {
		if err != nil {
			logger.Error(err, "")
		}
	}

	rootCmd.Flags().StringP("id", "i", "", "Sets the worker id (ID)")
	must(rootCmd.MarkFlagRequired("id"))

	_ = viper.BindPFlags(rootCmd.Flags())

	return rootCmd
}

// createViper creates a Viper object configured to read in configuration files
// (from various paths with content type specific filename extensions) and loads
// environment variables.
func createViper(logger logr.Logger, cmd *cobra.Command) error {
	viper.AutomaticEnv()
	viper.SetConfigName("virtual-worker")
	viper.AddConfigPath("/etc/tinkerbell")
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.Error(err, "could not load config file", "configFile", viper.ConfigFileUsed())
			return err
		}
		logger.Info("no config file found")
	} else {
		logger.Info("loaded config file", "configFile", viper.ConfigFileUsed())
	}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) {
			_ = cmd.Flags().SetAnnotation(f.Name, cobra.BashCompOneRequiredFlag, []string{"false"})
		}
	})

	return nil
}
