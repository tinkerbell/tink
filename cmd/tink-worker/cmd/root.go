package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/equinix-labs/otel-init-go/otelhelpers"
	"github.com/packethost/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-worker/internal"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc"
)

const (
	defaultRetryIntervalSeconds       = 3
	defaultRetryCount                 = 3
	defaultMaxFileSize          int64 = 10 * 1024 * 1024 // 10MB
	defaultTimeoutMinutes             = 60
)

// NewRootCommand creates a new Tink Worker Cobra root command.
func NewRootCommand(version string, logger log.Logger) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "tink-worker",
		Short:   "Tink Worker",
		Version: version,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			viper, err := createViper(logger)
			if err != nil {
				return err
			}
			return applyViper(viper, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			retryInterval, _ := cmd.Flags().GetDuration("retry-interval")
			retries, _ := cmd.Flags().GetInt("max-retry")
			// TODO(displague) is log-level no longer useful?
			// logLevel, _ := cmd.Flags().GetString("log-level")
			workerID, _ := cmd.Flags().GetString("id")
			maxFileSize, _ := cmd.Flags().GetInt64("max-file-size")
			timeOut, _ := cmd.Flags().GetDuration("timeout")
			user, _ := cmd.Flags().GetString("registry-username")
			pwd, _ := cmd.Flags().GetString("registry-password")
			registry, _ := cmd.Flags().GetString("docker-registry")
			captureActionLogs, _ := cmd.Flags().GetBool("capture-action-logs")

			logger.With("version", version).Info("starting")
			if setupErr := client.Setup(); setupErr != nil {
				return setupErr
			}

			ctx := context.Background()
			if timeOut > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeOut)
				defer cancel()
			}

			// if traceparent is set in /proc/cmdline or in the environment,
			// pick it up and add it to ctx
			ctx = otelhelpers.ContextWithCmdlineOrEnvTraceparent(ctx)

			conn, err := tryClientConnection(logger, retryInterval, retries)
			if err != nil {
				return err
			}
			rClient := pb.NewWorkflowServiceClient(conn)

			regConn := internal.NewRegistryConnDetails(registry, user, pwd, logger)
			worker := internal.NewWorker(rClient, regConn, logger, registry, retries, retryInterval, maxFileSize)

			err = worker.ProcessWorkflowActions(ctx, workerID, captureActionLogs)
			if err != nil {
				return errors.Wrap(err, "worker Finished with error")
			}
			return nil
		},
	}

	rootCmd.Flags().Duration("retry-interval", defaultRetryIntervalSeconds*time.Second, "Retry interval in seconds (RETRY_INTERVAL)")

	rootCmd.Flags().Duration("timeout", defaultTimeoutMinutes*time.Minute, "Max duration to wait for worker to complete (TIMEOUT)")

	rootCmd.Flags().Int("max-retry", defaultRetryCount, "Maximum number of retries to attempt (MAX_RETRY)")

	rootCmd.Flags().Int64("max-file-size", defaultMaxFileSize, "Maximum file size in bytes (MAX_FILE_SIZE)")

	rootCmd.Flags().Bool("capture-action-logs", true, "Capture action container output as part of worker logs")

	// rootCmd.Flags().String("log-level", "info", "Sets the worker log level (panic, fatal, error, warn, info, debug, trace)")

	must := func(err error) {
		if err != nil {
			logger.Fatal(err)
		}
	}

	rootCmd.Flags().StringP("id", "i", "", "Sets the worker id (ID)")
	must(rootCmd.MarkFlagRequired("id"))

	rootCmd.Flags().StringP("docker-registry", "r", "", "Sets the Docker registry (DOCKER_REGISTRY)")
	must(rootCmd.MarkFlagRequired("docker-registry"))

	rootCmd.Flags().StringP("registry-username", "u", "", "Sets the registry username (REGISTRY_USERNAME)")
	must(rootCmd.MarkFlagRequired("registry-username"))

	rootCmd.Flags().StringP("registry-password", "p", "", "Sets the registry-password (REGISTRY_PASSWORD)")
	must(rootCmd.MarkFlagRequired("registry-password"))

	return rootCmd
}

// createViper creates a Viper object configured to read in configuration files
// (from various paths with content type specific filename extensions) and loads
// environment variables.
func createViper(logger log.Logger) (*viper.Viper, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigName("tink-worker")
	v.AddConfigPath("/etc/tinkerbell")
	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.With("configFile", v.ConfigFileUsed()).Error(err, "could not load config file")
			return nil, err
		}
		logger.Info("no config file found")
	} else {
		logger.With("configFile", v.ConfigFileUsed()).Info("loaded config file")
	}

	return v, nil
}

func applyViper(v *viper.Viper, cmd *cobra.Command) error {
	errs := []error{}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				errs = append(errs, err)
				return
			}
		}
	})

	if len(errs) > 0 {
		es := []string{}
		for _, err := range errs {
			es = append(es, err.Error())
		}
		return fmt.Errorf(strings.Join(es, ", "))
	}

	return nil
}

func tryClientConnection(logger log.Logger, retryInterval time.Duration, retries int) (*grpc.ClientConn, error) {
	for ; retries > 0; retries-- {
		c, err := client.GetConnection()
		if err != nil {
			logger.With("error", err, "duration", retryInterval).Info("failed to connect, sleeping before retrying")
			<-time.After(retryInterval)
			continue
		}

		return c, nil
	}
	return nil, fmt.Errorf("retries exceeded")
}
