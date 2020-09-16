package template

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	tt "text/template"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/template"
)

var (
	fPath        = "path"
	fName        = "name"
	filePath     string
	templateName string
)

// createCmd represents the create subcommand for template command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a workflow template ",
	Example: `tink template create [flags]
cat /tmp/example.tmpl | tink template create -n example`,
	PreRunE: func(c *cobra.Command, args []string) error {
		if !isInputFromPipe() {
			path, _ := c.Flags().GetString(fPath)
			if path == "" {
				return fmt.Errorf("either pipe the template or provide the required '--path' flag")
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
			if err := tryParseTemplate(string(data)); err != nil {
				log.Fatal(err)
			}
			createTemplate(data)
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
	flags.StringVarP(&filePath, "path", "p", "", "path to the template file")
	flags.StringVarP(&templateName, "name", "n", "", "unique name for the template (alphanumeric)")
	_ = createCmd.MarkPersistentFlagRequired(fName)
}

func tryParseTemplate(data string) error {
	tmpl := *tt.New("")
	if _, err := tmpl.Parse(data); err != nil {
		return err
	}
	return nil
}

func createTemplate(data []byte) {
	req := template.WorkflowTemplate{Name: templateName, Data: string(data)}
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
