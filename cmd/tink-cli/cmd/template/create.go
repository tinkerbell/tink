package template

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/workflow"
)

var (
	file     = "file"
	filePath string
)

// createCmd represents the create subcommand for template command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a workflow template ",
	Example: `tink template create [flags]
cat /tmp/example.tmpl | tink template create -n example`,
	PreRunE: func(c *cobra.Command, args []string) error {
		if !isInputFromPipe() {
			path, _ := c.Flags().GetString(file)
			if path == "" {
				return errors.New("either pipe the template or provide the required '--file' flag")
			}
		}
		return nil
	},
	Run: func(c *cobra.Command, args []string) {
		var reader io.Reader
		if isInputFromPipe() {
			reader = os.Stdin
		} else {
			f, err := os.Open(filePath)
			if err != nil {
				log.Fatal(err)
			}
			reader = f
		}

		data := readAll(reader)
		if data != nil {
			wf, err := workflow.Parse(data)
			if err != nil {
				log.Fatal(err)
			}
			createTemplate(wf.Name, data)
		}
	},
}

func readAll(reader io.Reader) []byte {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func addFlags() {
	flags := createCmd.PersistentFlags()
	flags.StringVarP(&filePath, "file", "", "", "path to the template file")
}

func createTemplate(name string, data []byte) {
	req := template.WorkflowTemplate{Name: name, Data: string(data)}
	res, err := client.TemplateClient.CreateTemplate(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created Template: ", res.Id)
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func init() {
	addFlags()
	SubCommands = append(SubCommands, createCmd)
}
