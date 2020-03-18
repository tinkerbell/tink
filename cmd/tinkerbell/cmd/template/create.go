package template

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	tt "text/template"

	"github.com/packethost/tinkerbell/client"
	"github.com/packethost/tinkerbell/protos/template"
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
	Example: "tinkerbell template create [flags]",
	Run: func(c *cobra.Command, args []string) {
		validateTemplate()
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

func validateTemplate() {
	_, err := tt.ParseFiles(filePath)
	if err != nil {
		log.Fatalln(err)
	}
}

func readTemplateData() []byte {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func createTemplate(c *cobra.Command, args []string) {
	req := template.WorkflowTemplate{Name: templateName, Data: readTemplateData()}
	res, err := client.TemplateClient.CreateTemplate(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created Template: ", res.Id)
}

func init() {
	addFlags()
	SubCommands = append(SubCommands, createCmd)
}
