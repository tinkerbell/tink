package hardware

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/hardware"
	"github.com/spf13/cobra"
)

// allCmd represents the all command
var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Get all known hardware for facility",
	Run: func(cmd *cobra.Command, args []string) {
		alls, err := client.HardwareClient.All(context.Background(), &hardware.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		var hw *hardware.Hardware
		err = nil
		for hw, err = alls.Recv(); err == nil && hw != nil; hw, err = alls.Recv() {
			fmt.Println(hw.JSON)
		}
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
	},
}

func init() {
	SubCommands = append(SubCommands, allCmd)
}
