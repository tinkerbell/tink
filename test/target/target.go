package target

import (
	"context"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/target"
)

//var targetData = `{"targets": {"machine1": {"mac_addr": "98:03:9b:89:d7:ba"},"machine2": {"mac_addr": "98:03:9b:89:d7:ba"}}}`

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

func CreateTargets() (string, error) {
	i := int64(1)
	filepath := os.Getenv("GOPATH") + "/src/github.com/packethost/rover/test/target/data/target_" + strconv.FormatInt(i, 10) + ".json"
	data, err := getTargets(filepath)
	uuid, err := client.TargetClient.CreateTargets(context.Background(), &target.PushRequest{Data: data})
	if err != nil {
		return "", err
	}
	return uuid.Uuid, nil
}
