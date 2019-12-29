package framework

import (
	"os"
	"os/exec"
)

func startDb(filepath string) error {
	cmd := exec.Command("/bin/sh", "-c", "docker-compose -f "+filepath+" up --build -d db")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

// StartStack : Starting stack
func StartStack() error {
	// Docker compose file for starting the containers
	filepath := "../test-docker-compose.yml"

	// Start Db and logging components
	err := startDb(filepath)
	if err != nil {
		return err
	}

	// Start other containers
	cmd := exec.Command("/bin/sh", "-c", "docker-compose -f "+filepath+" up --build -d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return nil
}
