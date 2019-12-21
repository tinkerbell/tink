package e2e

import (
	"os"
	"os/exec"
	"time"
)

func startDb(filepath string) error {
	cmd := exec.Command("/bin/sh", "-c", "docker-compose -f "+filepath+" up --build -d db")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func startStack() error {
	// Docker compose file for starting the containers
	filepath := os.Getenv("GOPATH") + "src/github.com/packethost/rover/docker-compose.yml "

	// Start Db and logging components
	err := startDb(filepath)
	if err != nil {
		return err
	}

	// Wait for some time so thath the above containers to be in running condition
	time.Sleep(6 * time.Second)

	// Start other containers
	cmd := exec.Command("/bin/sh", "-c", "docker-compose -f "+filepath+" up --build -d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return nil
}
