package e2e

import (
	"os"
	"os/exec"
)

func tearDown() error {
	cmd := exec.Command("/bin/sh", "-c", "docker-compose rm -svf")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
