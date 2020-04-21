package framework

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/template"
)

func readTemplateData(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return []byte(""), err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return []byte(""), err
	}
	return data, nil
}

// CreateTemplate : create template in the database
func CreateTemplate(tmpl string) (string, error) {
	filePath := "data/template/" + tmpl
	// Read Content of template
	data, err := readTemplateData(filePath)
	req := template.WorkflowTemplate{Name: ("test_" + tmpl), Data: string(data)}
	res, err := client.TemplateClient.CreateTemplate(context.Background(), &req)
	if err != nil {
		return "", err
	}
	return res.Id, nil
}
