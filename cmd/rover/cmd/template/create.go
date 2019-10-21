package template

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/template"
	"github.com/spf13/cobra"
)

var (
	fPath        = "path"
	fName        = "name"
	filePath     string
	templateName string
)

// createCmd represents the create subcommand for template command
var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "create a workflow template ",
	Example: "rover template create [flags]",
	Run: func(c *cobra.Command, args []string) {
		createTemplate(c, args)
	},
}

func addFlags() {
	flags := createCmd.PersistentFlags()
	flags.StringVarP(&filePath, "path", "p", "", "path to the template file")
	flags.StringVarP(&templateName, "name", "n", "", "unique name for the template (alphanumeric)")

	createCmd.MarkPersistentFlagRequired(fPath)
	createCmd.MarkPersistentFlagRequired(fName)
}

func createTemplate(c *cobra.Command, args []string) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	req := template.WorkflowTemplate{Name: templateName, Data: data}
	id, err := client.TemplateClient.Create(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatalln(id)
}

func init() {
	addFlags()
	SubCommands = append(SubCommands, createCmd)
}
