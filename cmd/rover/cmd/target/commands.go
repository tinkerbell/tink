package target

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

// SubCommands holds the sub commands for targets command
// Example: rover targets [subcommand]
var SubCommands []*cobra.Command

func verifyUUIDs(args []string) error {
	if len(args) < 1 {
		return errors.New("requires at least one id")
	}
	for _, arg := range args {
		if _, err := uuid.FromString(arg); err != nil {
			return fmt.Errorf("invalid uuid: %s", arg)
		}
	}
	return nil
}

type machine map[string]string

type tmap struct {
	Targets map[string]machine `json:"targets"`
}

func isValidData(arg []byte) error {
	var tr tmap
	if json.Unmarshal([]byte(arg), &tr) != nil {
		return fmt.Errorf("invalid json: %s", arg)
	}
	for _, v := range tr.Targets {
		for key, val := range v {
			switch key {
			case string("mac_addr"):
				_, err := net.ParseMAC(val)
				if err != nil {
					return err
				}
			case string("ipv4_addr"):
				ip := net.ParseIP(val)
				if ip == nil || ip.To4() == nil {
					return fmt.Errorf("invalid ip_addr: %s", val)
				}
			default:
				return fmt.Errorf("invalid key \"%s\" in data. it should be \"mac_addr\" or \"ip_addr\" only", key)
			}
		}
	}
	return nil
}
