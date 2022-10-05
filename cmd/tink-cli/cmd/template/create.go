//nolint:import-shadowing // The name 'template' shadows an import name
package template

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/workflow"
)

var filePath string

func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a workflow template ",
		Long: `The create command allows you create workflow templates:
# Pipe the file to create a template:
$ cat /tmp/example.tmpl | tink template create
# Create template using the --file flag:
$ tink template create --file /tmp/example.tmpl
`,
		PreRunE: func(c *cobra.Command, args []string) error {
			if !isInputFromPipe() {
				if filePath == "" {
					return fmt.Errorf("%v requires the '--file' flag", c.UseLine())
				}
			}
			return nil
		},
		Run: func(c *cobra.Command, args []string) {
			var reader io.Reader
			if isInputFromPipe() {
				reader = os.Stdin
			} else {
				f, err := os.Open(filepath.Clean(filePath))
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
	flags := cmd.PersistentFlags()
	flags.StringVar(&filePath, "file", "./template.yaml", "path to the template file")
	return cmd
}

func readAll(reader io.Reader) []byte {
	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	return data
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
