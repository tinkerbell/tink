package framework

import (
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

func removeWorkerImage() error {
	cmd := exec.Command("/bin/sh", "-c", "docker image rm worker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func createWorkerImage() error {
	cmd := exec.Command("/bin/sh", "-c", "docker build -t worker ../worker/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		logger.Errorln("Failed to create worker image", err)
	}
	logger.Infoln("Worker Image created")
	return err
}

func initializeLogger() {
	level := os.Getenv("TEST_LOG_LEVEL")
	if level != "" {
		switch strings.ToLower(level) {
		case "panic":
			logger.SetLevel(logrus.PanicLevel)
		case "fatal":
			logger.SetLevel(logrus.FatalLevel)
		case "error":
			logger.SetLevel(logrus.ErrorLevel)
		case "warn", "warning":
			logger.SetLevel(logrus.WarnLevel)
		case "info":
			logger.SetLevel(logrus.InfoLevel)
		case "debug":
			logger.SetLevel(logrus.DebugLevel)
		case "trace":
			logger.SetLevel(logrus.TraceLevel)
		default:
			logger.SetLevel(logrus.InfoLevel)
			logger.Errorln("Invalid value for TEST_LOG_LEVEL ", level, " .Setting it to default(Info)")
		}
	} else {
		logger.SetLevel(logrus.InfoLevel)
		logger.Errorln("Variable TEST_LOG_LEVEL is not set. Default is Info.")
	}
	logger.SetFormatter(&logrus.JSONFormatter{})
}

// StartStack : Starting stack.
func StartStack() error {
	// Docker compose file for starting the containers
	filepath := "../test-docker-compose.yml"

	// Initialize logger
	initializeLogger()

	// Start Db and logging components
	err := startDb(filepath)
	if err != nil {
		return err
	}

	// Building certs
	err = buildCerts(filepath)
	if err != nil {
		return err
	}

	// Building registry
	err = buildLocalDockerRegistry(filepath)
	if err != nil {
		return err
	}

	// Build default images
	err = buildActionImages()
	if err != nil {
		return err
	}

	// Push Images into registry
	err = pushImages()
	if err != nil {
		return err
	}

	// Remove older worker image
	err = removeWorkerImage()
	if err != nil {
		return err
	}

	// Create new Worker image locally
	err = createWorkerImage()
	if err != nil {
		logger.Errorln("failed to create worker Image")
		return errors.Wrap(err, "worker image creation failed")
	}

	initializeLogger()

	// Start other containers
	cmd := exec.Command("/bin/sh", "-c", "docker-compose -f "+filepath+" up --build -d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		logger.Errorln("failed to create worker Image")
		return errors.Wrap(err, "worker image creation failed")
	}
	return nil
}
