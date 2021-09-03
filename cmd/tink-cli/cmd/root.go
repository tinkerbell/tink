package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tinkerbell/tink/client"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:               "tink",
	Short:             "tinkerbell CLI",
	PersistentPreRunE: setupClient,
	DisableAutoGenTag: true,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "facility", "f", "", "used to build grpc and http urls")
}

func setupClient(_ *cobra.Command, _ []string) error {
	return client.Setup()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) error {
	rootCmd.Version = version
	rootCmd.AddCommand(NewHardwareCommand())
	rootCmd.AddCommand(NewTemplateCommand())
	rootCmd.AddCommand(NewWorkflowCommand())
	return rootCmd.Execute()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
