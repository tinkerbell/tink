// Copyright © 2018 packet.net

package hardware

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/hardware"
)

var (
	filePath string
	fPath    = "file-path"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push new hardware to tinkerbell",
	Example: `cat /tmp/data.json | tink hardware push
tink hardware push -p /tmp/data.json`,
	PreRunE: func(c *cobra.Command, args []string) error {
		if !isInputFromPipe() {
			path, _ := c.Flags().GetString(fPath)
			if path == "" {
				return fmt.Errorf("%v either pipe the data or provide the required flag", c.UseLine())
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var data string
		if isInputFromPipe() {
			data = readDataFromStdin()
		} else {
			data = readDataFromFile()
		}
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

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func readDataFromStdin() string {
	scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
	for scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func readDataFromFile() string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}

func init() {
	flags := pushCmd.PersistentFlags()
	flags.StringVarP(&filePath, "file-path", "p", "", "path to the hardware data file")

	SubCommands = append(SubCommands, pushCmd)
}
