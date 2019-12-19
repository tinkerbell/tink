package template

import (
	"context"
	"io/ioutil"
	"os"
	"strconv"

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

func CreateTemplate() (string, error) {
	i := int64(1)
	filepath := os.Getenv("GOPATH") + "/src/github.com/packethost/rover/test/template/data/sample_" + strconv.FormatInt(i, 10) + ".tmpl"
	data, err := readTemplateData(filepath)
	req := template.WorkflowTemplate{Name: ("test_template_" + strconv.FormatInt(i, 10)), Data: data}
	res, err := client.TemplateClient.CreateTemplate(context.Background(), &req)
	if err != nil {
		return "", err
	}
	return res.Id, nil
}
