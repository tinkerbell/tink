package framework

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/template"
)

func readTemplateData(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func() {
		f.Close()
	}()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CreateTemplate : create template in the database
func CreateTemplate(tmpl string) (string, error) {
	filePath := "data/template/" + tmpl
	// Read Content of template
	data, err := readTemplateData(filePath)
	if err != nil {
		return "", err
	}
	req := template.WorkflowTemplate{Name: ("test_" + tmpl), Data: data}
	res, err := client.TemplateClient.CreateTemplate(context.Background(), &req)
	if err != nil {
		return "", err
	}
	return res.Id, nil
}
