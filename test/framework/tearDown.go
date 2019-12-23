package framework

import (
	"os"
	"os/exec"
)

func TearDown() error {
	cmd := exec.Command("/bin/sh", "-c", "docker-compose rm -svf")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
