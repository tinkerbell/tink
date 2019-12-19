package hardware

import (
	"context"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/hardware"
)

func readHwData(file string) (string, error) {
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

func PushHardwareData() error {
	i := int64(1)
	filepath := os.Getenv("GOPATH") + "/src/github.com/packethost/rover/test/hardware/data/hardware_" + strconv.FormatInt(i, 10) + ".json"
	data, err := readHwData(filepath)
	if err != nil {
		return err
	}
	_, err = client.HardwareClient.Push(context.Background(), &hardware.PushRequest{Data: data})
	return err
}
