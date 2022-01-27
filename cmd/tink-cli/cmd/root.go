package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/internal/clientctx"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:               "tink",
	Short:             "tinkerbell CLI",
	DisableAutoGenTag: true,
}

// gets passed into to subcommands as a pointer in the context.Context.
var fullClient = client.FullClient{}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringP("facility", "f", "", "used to build grpc and http urls")
	rootCmd.PersistentFlags().String("tinkerbell-grpc-authority", "127.0.0.1:42113", "Connection info for tink-server")
	rootCmd.PersistentFlags().Bool("tinkerbell-tls", true, "Connect to server via TLS or not")
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) error {
	opts := client.ConnOptions{
		GRPCAuthority: viper.GetString("tinkerbell-grpc-authority"),
		TLS:           viper.GetBool("tinkerbell-tls"),
	}
	conn, err := client.NewClientConn(&opts)
	if err != nil {
		return err
	}
	client.HardwareClient = hardware.NewHardwareServiceClient(conn)
	client.TemplateClient = template.NewTemplateServiceClient(conn)
	client.WorkflowClient = workflow.NewWorkflowServiceClient(conn)
	fullClient = client.FullClient{
		HardwareClient: client.HardwareClient,
		TemplateClient: client.TemplateClient,
		WorkflowClient: client.WorkflowClient,
	}

	rootCmd.Version = version
	rootCmd.AddCommand(NewHardwareCommand())
	rootCmd.AddCommand(NewTemplateCommand())
	rootCmd.AddCommand(NewWorkflowCommand())

	ctx := clientctx.Set(context.Background(), &fullClient)
	return rootCmd.ExecuteContext(ctx)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
