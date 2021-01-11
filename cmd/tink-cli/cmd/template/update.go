package template

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/workflow"
)

// updateCmd represents the get subcommand for template command
func NewUpdateCommand(cl *client.MetaClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [id] [flags]",
		Short: "update a workflow template",
		Long: `The update command allows you change the definition of an existing workflow template:
# Update an existing template:
$ tink template update 614168df-45a5-11eb-b13d-0242ac120003 --file /tmp/example.tmpl
`,
		PreRunE: func(c *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("%v requires the '--file' flag", c.UseLine())
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
				updateTemplate(cl, arg)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&filePath, "path", "p", "", "path to the template file")
	return cmd
}

func updateTemplate(cl *client.MetaClient, id string) {
	req := template.WorkflowTemplate{Id: id}
	if filePath != "" {
		data := readTemplateData()
		if data != "" {
			wf, err := workflow.Parse([]byte(data))
			if err != nil {
				log.Fatal(err)
			}
			req.Name = wf.Name
			req.Data = data
		}
	} else {
		log.Fatal("Nothing is provided in the file path")
	}

	_, err := cl.TemplateClient.UpdateTemplate(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated Template: ", id)
}

func readTemplateData() string {
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
