package template

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/template"
)

// updateCmd represents the get subcommand for template command
var updateCmd = &cobra.Command{
	Use:     "update [id] [flags]",
	Short:   "update a template",
	Example: "tink template update [id] [flags]",
	PreRunE: func(c *cobra.Command, args []string) error {
		name, _ := c.Flags().GetString(fName)
		path, _ := c.Flags().GetString(fPath)
		if name == "" && path == "" {
			return fmt.Errorf("%v requires at least one flag", c.UseLine())
		}
		return nil
	},
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires argument", c.UseLine())
		}
		for _, arg := range args {
			if _, err := uuid.FromString(arg); err != nil {
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

func updateTemplate(id string) {
	req := template.WorkflowTemplate{Id: id}
	if filePath == "" && templateName != "" {
		req.Name = templateName
	} else if filePath != "" && templateName == "" {
		data := readTemplateData()
		if data != "" {
			if err := tryParseTemplate(data); err != nil {
				log.Println(err)
				return
			}
			req.Data = data
		}
	} else {
		req.Name = templateName
		req.Data = readTemplateData()
	}

	_, err := client.TemplateClient.UpdateTemplate(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated Template: ", id)
}

func readTemplateData() string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
	}
	return string(data)
}

func init() {
	flags := updateCmd.PersistentFlags()
	flags.StringVarP(&filePath, "path", "p", "", "path to the template file")
	flags.StringVarP(&templateName, "name", "n", "", "unique name for the template (alphanumeric)")

	SubCommands = append(SubCommands, updateCmd)
}
