// Copyright © 2018 packet.net

package hardware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/hardware"
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:     "push",
	Short:   "Push new hardware to rover",
	Example: `rover hardware push '{"id":"2a1519e5-781c-4251-a979-3a6bedb8ba59", ...}' '{"id:"315169a4-a863-43ef-8817-2b6a57bd1eef", ...}'`,
	Args: func(_ *cobra.Command, args []string) error {
		s := struct {
			ID string
		}{}
		for _, arg := range args {
			if json.NewDecoder(strings.NewReader(arg)).Decode(&s) != nil {
				return fmt.Errorf("invalid json: %s", arg)
			} else if s.ID == "" {
				return fmt.Errorf("invalid json, ID is required: %s", arg)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, j := range args {
			if _, err := client.HardwareClient.Push(context.Background(), &hardware.PushRequest{Data: j}); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	SubCommands = append(SubCommands, pushCmd)
}
