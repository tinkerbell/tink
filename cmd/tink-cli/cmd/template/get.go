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

// getCmd represents the get subcommand for template command
var getCmd = &cobra.Command{
	Use:     "get [id]",
	Short:   "get a template",
	Example: "tink template get [id]",
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
			t, err := client.TemplateClient.GetTemplate(context.Background(), &req)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(t.Data))
		}
	},
}

func init() {
	getCmd.DisableFlagsInUseLine = true
	SubCommands = append(SubCommands, getCmd)
}
