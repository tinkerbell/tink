package hardware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/hardware"
)

// allCmd represents the all command
var allCmd = &cobra.Command{
	Use:   "all",
	Short: "get all known hardware for facility",
	Run: func(cmd *cobra.Command, args []string) {
		alls, err := client.HardwareClient.All(context.Background(), &hardware.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		var hw *hardware.Hardware
		err = nil
		for hw, err = alls.Recv(); err == nil && hw != nil; hw, err = alls.Recv() {
			b, err := json.Marshal(hw)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(b))
		}
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
	},
}

func init() {
	SubCommands = append(SubCommands, allCmd)
}
