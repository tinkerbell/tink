package target

import (
	"context"
	"fmt"
	"log"

	"github.com/packethost/tinkerbell/client"
	"github.com/packethost/tinkerbell/protos/target"
	"github.com/spf13/cobra"
)

var (
	id       = "uuid"
	data     = "jsonData"
	uid      string
	jsonData string
)

// createCmd represents the create sub command for template command
var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "update a target",
	Example: "tinkerbell target update [flags]",
	Run: func(c *cobra.Command, args []string) {
		err := validateData(c, args)
		if err != nil {
			log.Fatal(err)
		}
		updateTargets(c, args)
	},
}

func addFlags() {
	flags := updateCmd.PersistentFlags()
	flags.StringVarP(&uid, "uuid", "u", "", "id for targets to be updated")
	flags.StringVarP(&jsonData, "jsondata", "j", "", "JSON data which needs to be pushed")
	updateCmd.MarkPersistentFlagRequired(id)
	updateCmd.MarkPersistentFlagRequired(data)
}

func validateData(c *cobra.Command, args []string) error {
	err := isValidData([]byte(jsonData))
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
func updateTargets(c *cobra.Command, args []string) {
	if _, err := client.TargetClient.UpdateTargetByID(context.Background(), &target.UpdateRequest{ID: uid, Data: jsonData}); err != nil {
		log.Fatal(err)
	}
}

func init() {
	addFlags()
	SubCommands = append(SubCommands, updateCmd)
}
