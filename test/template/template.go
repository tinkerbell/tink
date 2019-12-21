package template

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/template"
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

func CreateTemplate(tmpl string) (string, error) {
	filepath := os.Getenv("GOPATH") + "/src/github.com/packethost/rover/test/template/data/" + tmpl
	data, err := readTemplateData(filepath)
	req := template.WorkflowTemplate{Name: ("test_" + tmpl), Data: data}
	res, err := client.TemplateClient.CreateTemplate(context.Background(), &req)
	if err != nil {
		return "", err
	}
	return res.Id, nil
}
