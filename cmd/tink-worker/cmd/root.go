package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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
	retryIntervalDefault        = 3
	retryCountDefault           = 3
	defaultMaxFileSize    int64 = 10485760 //10MB ~= 10485760Bytes
	defaultTimeoutMinutes       = 60
)

// NewRootCommand creates a new Tink Worker Cobra root command
func NewRootCommand(version string, logger log.Logger) *cobra.Command {
	must := func(err error) {
		if err != nil {
			logger.Fatal(err)
		}
	}

	rootCmd := &cobra.Command{
		Use:     "tink-worker",
		Short:   "Tink Worker",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper, err := createViper()
			if err != nil {
				return err
			}
			return applyViper(viper, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			retryInterval, _ := cmd.PersistentFlags().GetDuration("retry-interval")
			retries, _ := cmd.PersistentFlags().GetInt("retries")
			// TODO(displague) is log-level no longer useful?
			// logLevel, _ := cmd.PersistentFlags().GetString("log-level")
			workerID, _ := cmd.PersistentFlags().GetString("id")
			maxFileSize, _ := cmd.PersistentFlags().GetInt64("max-file-size")
			timeOut, _ := cmd.PersistentFlags().GetDuration("timeout")
			user, _ := cmd.PersistentFlags().GetString("registry-username")
			pwd, _ := cmd.PersistentFlags().GetString("registry-password")
			registry, _ := cmd.PersistentFlags().GetString("docker-registry")

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

			conn, err := tryClientConnection(logger, retryInterval, retries)
			if err != nil {
				return err
			}
			rClient := pb.NewWorkflowSvcClient(conn)

			regConn := internal.NewRegistryConnDetails(registry, user, pwd, logger)
			worker := internal.NewWorker(rClient, regConn, logger, registry, retries, retryInterval, maxFileSize)

			err = worker.ProcessWorkflowActions(ctx, workerID)
			if err != nil {
				return errors.Wrap(err, "worker Finished with error")
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().Duration("retry-interval", retryIntervalDefault, "Retry interval in seconds")

	rootCmd.PersistentFlags().Duration("timeout", time.Duration(defaultTimeoutMinutes*time.Minute), "Max duration to wait for worker to complete")

	rootCmd.PersistentFlags().Int("max-retry", retryCountDefault, "Maximum number of retries to attempt")

	rootCmd.PersistentFlags().Int64("max-file-size", defaultMaxFileSize, "Maximum file size in bytes")

	// rootCmd.PersistentFlags().String("log-level", "info", "Sets the worker log level (panic, fatal, error, warn, info, debug, trace)")

	rootCmd.PersistentFlags().StringP("id", "i", "", "Sets the worker id")
	must(rootCmd.MarkPersistentFlagRequired("id"))

	rootCmd.PersistentFlags().StringP("docker-registry", "r", "", "Sets the Docker registry")
	must(rootCmd.MarkPersistentFlagRequired("docker-registry"))

	rootCmd.PersistentFlags().StringP("registry-username", "u", "", "Sets the registry username")
	must(rootCmd.MarkPersistentFlagRequired("registry-username"))

	rootCmd.PersistentFlags().StringP("registry-password", "p", "", "Sets the registry-password")
	must(rootCmd.MarkPersistentFlagRequired("registry-password"))

	return rootCmd
}

// createViper creates a Viper object configured to read in configuration files
// (from various paths with content type specific filename extensions) and load
// environment variables that start with TINK_WORKER.
func createViper() (*viper.Viper, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigName("tink-worker")
	v.AddConfigPath("/etc/tinkerbell")
	v.AddConfigPath(".")
	v.SetEnvPrefix("TINK_WORKER")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	} else {
		fmt.Fprintln(os.Stderr, "Using config file:", v.ConfigFileUsed())
	}

	return v, nil
}

func applyViper(v *viper.Viper, cmd *cobra.Command) error {
	errors := []error{}

	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				errors = append(errors, err)
				return
			}
		}
	})

	if len(errors) > 0 {
		errs := []string{}
		for _, err := range errors {
			errs = append(errs, err.Error())
		}
		return fmt.Errorf(strings.Join(errs, ", "))
	}

	return nil
}

func tryClientConnection(logger log.Logger, retryInterval time.Duration, retries int) (*grpc.ClientConn, error) {
	for ; retries > 0; retries-- {
		c, err := client.GetConnection()
		if err != nil {
			logger.With("error", err, "duration", retryInterval).Info("failed to connect, sleeping before retrying")
			<-time.After(retryInterval * time.Second)
			continue
		}

		return c, nil
	}
	return nil, fmt.Errorf("retries exceeded")
}
