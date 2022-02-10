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
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		// we have to connect to tink server here and not any earlier because cobra
		// would not have run yet and thus hasn't parsed the cli flags which would
		// override env or config file
		conn, err := client.NewClientConn(
			viper.GetString("tinkerbell-grpc-authority"),
			viper.GetBool("tinkerbell-tls"),
		)
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

		return nil
	},
}

// gets passed into to subcommands as a pointer in the context.Context.
var fullClient = client.FullClient{}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) error {
	rootCmd.Version = version
	rootCmd.AddCommand(NewHardwareCommand())
	rootCmd.AddCommand(NewTemplateCommand())
	rootCmd.AddCommand(NewWorkflowCommand())

	rootCmd.PersistentFlags().StringP("facility", "f", "", "used to build grpc and http urls")
	rootCmd.PersistentFlags().String("tinkerbell-grpc-authority", "", "Connection info for tink-server (TINKERBELL_GRPC_AUTHORITY)")
	rootCmd.PersistentFlags().Bool("tinkerbell-tls", true, "Connect to server via TLS or not")

	// Both AutomaticEnv and SetEnvKeyReplacer need to be called/setup before the VisitAll command that follows, otherwise env vars don't count
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	_ = viper.BindPFlags(rootCmd.PersistentFlags())

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	ctx := clientctx.Set(context.Background(), &fullClient)
	return rootCmd.ExecuteContext(ctx)
}
