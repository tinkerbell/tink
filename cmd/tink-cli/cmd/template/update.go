package template

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/workflow"
)

// updateCmd represents the get subcommand for template command.
func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [id] [flags]",
		Short: "update a workflow template",
		Long: `The update command allows you change the definition of an existing workflow template:
# Pipe the file to create a template:
$ cat /tmp/example.tmpl | tink template update 614168df-45a5-11eb-b13d-0242ac120003
# Update an existing template:
$ tink template update 614168df-45a5-11eb-b13d-0242ac120003 --path /tmp/example.tmpl
`,
		PreRunE: func(c *cobra.Command, args []string) error {
			if !isInputFromPipe() {
				if filePath == "" {
					return fmt.Errorf("%v requires the '--path' flag", c.UseLine())
				}
			}
			return nil
		},
		Args: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("%v requires argument", c.UseLine())
			}
			for _, arg := range args {
				if _, err := uuid.Parse(arg); err != nil {
					return fmt.Errorf("invalid uuid: %s", arg)
				}
			}
			return nil
		},
		Run: func(c *cobra.Command, args []string) {
			for _, arg := range args {
				updateTemplate(arg)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&filePath, "path", "p", "", "path to the template file")
	return cmd
}

func updateTemplate(id string) {
	req := template.WorkflowTemplate{Id: id}
	data, err := readTemplateData()
	if err != nil {
		log.Fatalf("readTemplateData: %v", err)
	}

	if data != "" {
		wf, err := workflow.Parse([]byte(data))
		if err != nil {
			log.Fatal(err)
		}
		req.Name = wf.Name
		req.Data = data
	}

	if _, err := client.TemplateClient.UpdateTemplate(context.Background(), &req); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated Template: ", id)
}

func readTemplateData() (string, error) {
	var reader io.Reader
	if isInputFromPipe() {
		reader = os.Stdin
	} else {
		f, err := os.Open(filepath.Clean(filePath))
		if err != nil {
			return "", err
		}

		defer f.Close()

		reader = f
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
