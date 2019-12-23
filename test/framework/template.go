package framework

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
	//matches, err := filepath.Glob("*/" + tmpl + "*")
	//var filePath string
	filePath := os.Getenv("GOPATH") + "/src/github.com/packethost/rover/test/data/template/" + tmpl
	/*if err != nil {
		filePath = matches[0]
	} else {
		fmt.Println("Match not found ", matches, " for pattern ", tmpl)
		return "", err
	}*/
	//fmt.Println("Reading template : ", filePath)
	data, err := readTemplateData(filePath)
	req := template.WorkflowTemplate{Name: ("test_" + tmpl), Data: data}
	res, err := client.TemplateClient.CreateTemplate(context.Background(), &req)
	if err != nil {
		return "", err
	}
	return res.Id, nil
}
