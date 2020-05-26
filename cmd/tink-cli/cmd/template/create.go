package template

import (
	"context"
	"fmt"
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
		var data []byte
		if isInputFromPipe() {
			data = readDataFromStdin()
		} else {
			data = readTemplateData()
		}
		if data != nil {
			if err := tryParseTemplate(data); err != nil {
				log.Println(err)
				return
			}
			createTemplate(c, data)
		}
	},
}

func addFlags() {
	flags := createCmd.PersistentFlags()
	flags.StringVarP(&filePath, "path", "p", "", "path to the template file")
	flags.StringVarP(&templateName, "name", "n", "", "unique name for the template (alphanumeric)")
	createCmd.MarkPersistentFlagRequired(fName)
}

func tryParseTemplate(data []byte) error {
	tmpl := *tt.New("")
	if _, err := tmpl.Parse(string(data)); err != nil {
		return err
	}
	return nil
}

func readTemplateData() []byte {
	f, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
	}
	return data
}

func createTemplate(c *cobra.Command, data []byte) {
	req := template.WorkflowTemplate{Name: templateName, Data: data}
	res, err := client.TemplateClient.CreateTemplate(context.Background(), &req)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("Created Template: ", res.Id)
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func readDataFromStdin() []byte {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil
	}
	return data
}

func init() {
	addFlags()
	SubCommands = append(SubCommands, createCmd)
}
