// Copyright © 2018 packet.net

package hardware

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/hardware"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:     "push",
	Short:   "Push new hardware to tinkerbell",
	Example: "cat data.json | tink hardware push",
	Run: func(cmd *cobra.Command, args []string) {
		data := readHardwareData(os.Stdin)
		s := struct {
			ID string
		}{}
		if json.NewDecoder(strings.NewReader(data)).Decode(&s) != nil {
			log.Fatalf("invalid json: %s", data)
		} else if s.ID == "" {
			log.Fatalf("invalid json, ID is required: %s", data)
		}
		if _, err := client.HardwareClient.Push(context.Background(), &hardware.PushRequest{Data: data}); err != nil {
			log.Fatal(err)
		}
		log.Println("Hardware data pushed successfully")
	},
}

func readHardwareData(r io.Reader) string {
	scanner := bufio.NewScanner(bufio.NewReader(r))
	for scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func init() {
	SubCommands = append(SubCommands, pushCmd)
}
