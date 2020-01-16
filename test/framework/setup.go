package framework

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

func buildCerts(filepath string) error {
	cmd := exec.Command("/bin/sh", "-c", "docker-compose -f "+filepath+" up --build certs")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func buildLocalDockerRegistry(filepath string) error {
	cmd := exec.Command("/bin/sh", "-c", "docker-compose -f "+filepath+" up --build -d registry")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func buildActionImages() error {
	cmd := exec.Command("/bin/sh", "-c", "./build_images.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func pushImages() error {
	cmd := exec.Command("/bin/sh", "-c", "./push_images.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

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

	// Building certs
	err := buildCerts(filepath)
	if err != nil {
		return err
	}

	// Building registry
	err = buildLocalDockerRegistry(filepath)
	if err != nil {
		return err
	}

	//Build default images
	err = buildActionImages()
	if err != nil {
		return err
	}

	//Push Images into registry
	err = pushImages()
	if err != nil {
		return err
	}

	// Start Db and logging components
	err = startDb(filepath)
	if err != nil {
		return err
	}

	//Create Worker image locally
	err = createWorkerImage()
	if err != nil {
		fmt.Println("failed to create worker Image")
		return errors.Wrap(err, "worker image creation failed")
	}

	// Start other containers
	cmd := exec.Command("/bin/sh", "-c", "docker-compose -f "+filepath+" up --build -d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return nil
}
