package template

import (
	"context"
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/template"
)

// deleteCmd represents the delete subcommand for template command
var deleteCmd = &cobra.Command{
	Use:     "delete [id]",
	Short:   "delete a template",
	Example: "tink template delete [id]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires an argument", c.UseLine())
		}
		for _, arg := range args {
			if _, err := uuid.FromString(arg); err != nil {
				return fmt.Errorf("invalid uuid: %s", arg)
			}
		}
		return nil
	},
	Run: func(c *cobra.Command, args []string) {
		for _, arg := range args {
			req := template.GetRequest{Id: arg}
			if _, err := client.TemplateClient.DeleteTemplate(context.Background(), &req); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	deleteCmd.DisableFlagsInUseLine = true
	SubCommands = append(SubCommands, deleteCmd)
}
