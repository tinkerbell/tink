package framework

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/packethost/tinkerbell/client"
	"github.com/packethost/tinkerbell/protos/hardware"
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

// PushHardwareData : push hardware data
func PushHardwareData(hwDataFiles []string) error {
	for _, hwFile := range hwDataFiles {
		filepath := "data/hardware/" + hwFile
		data, err := readHwData(filepath)
		if err != nil {
			return err
		}
		_, err = client.HardwareClient.Push(context.Background(), &hardware.PushRequest{Data: data})
		if err != nil {
			return err
		}
	}
	return nil
}
