package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/packethost/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tinkerbell/tink/pkg/controllers"
	wfctrl "github.com/tinkerbell/tink/pkg/controllers/workflow"
	"k8s.io/client-go/tools/clientcmd"
)

// version is set at build time.
var version = "devel"

// DaemonConfig represents all the values you can configure as part of the tink-server.
// You can change the configuration via environment variable, or file, or command flags.
type DaemonConfig struct {
	K8sAPI     string
	Kubeconfig string // only applies to out of cluster
}

func (c *DaemonConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.K8sAPI, "kubernetes", "", "The Kubernetes URL")
	fs.StringVar(&c.Kubeconfig, "kubeconfig", "", "Absolute path to the kubeconfig file")
}

func main() {
	logger, err := log.Init("github.com/tinkerbell/tink")
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	config := &DaemonConfig{}

	cmd := NewRootCommand(config, logger)
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		defer os.Exit(1)
	}
}

func NewRootCommand(config *DaemonConfig, logger log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use: "tink-controller",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			viper, err := createViper(logger)
			if err != nil {
				return err
			}
			return applyViper(viper, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("starting controller version " + version)

			config, err := clientcmd.BuildConfigFromFlags(config.K8sAPI, config.Kubeconfig)
			if err != nil {
				return err
			}

			manager, err := controllers.NewManager(config, controllers.GetControllerOptions())
			if err != nil {
				return err
			}

			return manager.RegisterControllers(
				cmd.Context(),
				wfctrl.NewController(manager.GetClient()),
			).Start(cmd.Context())
		},
	}
	config.AddFlags(cmd.Flags())
	return cmd
}

func createViper(logger log.Logger) (*viper.Viper, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigName("tink-controller")
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
	errors := []error{}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
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
