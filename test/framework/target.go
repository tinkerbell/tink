package framework

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/target"
)

func getTargets(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CreateTargets : create target in the database
func CreateTargets(tar string) (string, error) {
	filepath := "data/target/" + tar
	data, err := getTargets(filepath)
	uuid, err := client.TargetClient.CreateTargets(context.Background(), &target.PushRequest{Data: data})
	if err != nil {
		return "", err
	}
	return uuid.Uuid, nil
}
