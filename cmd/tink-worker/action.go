package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

var (
	registry string
	cli      *client.Client
	log      *logrus.Entry
)

func executeAction(ctx context.Context, action *pb.WorkflowAction, wfID string) (string, pb.ActionState, error) {
	log = logger.WithFields(logrus.Fields{"workflow_id": wfID, "worker_id": action.GetWorkerId()})
	err := pullActionImage(ctx, action)
	if err != nil {
		return fmt.Sprintf("Failed to pull Image : %s", action.GetImage()), 1, errors.Wrap(err, "DOCKER PULL")
	}
	id, err := createContainer(ctx, action, action.Command, wfID)
	if err != nil {
		return "Failed to create container", 1, errors.Wrap(err, "DOCKER CREATE")
	}
	var timeCtx context.Context
	var cancel context.CancelFunc
	if action.Timeout > 0 {
		timeCtx, cancel = context.WithTimeout(context.Background(), time.Duration(action.Timeout)*time.Second)
	} else {
		timeCtx, cancel = context.WithTimeout(context.Background(), 1*time.Hour)
	}
	defer cancel()
	//run container with timeout context
	//startedAt := time.Now()
	err = runContainer(timeCtx, id)
	if err != nil {
		return "Failed to run container", 1, errors.Wrap(err, "DOCKER RUN")
	}

	failedActionStatus := make(chan pb.ActionState)

	//capturing logs of action container in a go-routine
	go captureLogs(ctx, id)

	status, err, werr := waitContainer(timeCtx, id)
	if werr != nil {
		rerr := removeContainer(ctx, id)
		if rerr != nil {
			log.WithField("container_id", id).Errorln("Failed to remove container as ", rerr)
		}
		return "Failed to wait for completion of action", status, errors.Wrap(err, "DOCKER_WAIT")
	}
	rerr := removeContainer(ctx, id)
	if rerr != nil {
		return "Failed to remove container of action", status, errors.Wrap(rerr, "DOCKER_REMOVE")
	}
	log.Infoln("Container removed with Status ", pb.ActionState(status))
	if status != pb.ActionState_ACTION_SUCCESS {
		if status == pb.ActionState_ACTION_TIMEOUT && action.OnTimeout != nil {
			id, err = createContainer(ctx, action, action.OnTimeout, wfID)
			if err != nil {
				log.Errorln("Failed to create container for on-timeout command: ", err)
			}
			log.Infoln("Container created with on-timeout command : ", action.OnTimeout)
			failedActionStatus := make(chan pb.ActionState)
			go captureLogs(ctx, id)
			go waitFailedContainer(ctx, id, failedActionStatus)
			err = runContainer(ctx, id)
			if err != nil {
				log.Errorln("Failed to run on-timeout command: ", err)
			}
			onTimeoutStatus := <-failedActionStatus
			log.Infoln("On-Timeout Container status : ", onTimeoutStatus)
		} else {
			if action.OnFailure != nil {
				id, err = createContainer(ctx, action, action.OnFailure, wfID)
				if err != nil {
					log.Errorln("Failed to create on-failure command: ", err)
				}
				log.Infoln("Container created with on-failure command : ", action.OnFailure)
				go captureLogs(ctx, id)
				go waitFailedContainer(ctx, id, failedActionStatus)
				err = runContainer(ctx, id)
				if err != nil {
					log.Errorln("Failed to run on-failure command: ", err)
				}
				onFailureStatus := <-failedActionStatus
				log.Infoln("on-failure Container status : ", onFailureStatus)
			}
		}
		log.Infoln("Wait finished for failed or timeout container")
		if err != nil {
			rerr := removeContainer(ctx, id)
			if rerr != nil {
				log.Errorln("Failed to remove container as ", rerr)
			}
			log.Infoln("Failed to wait for container : ", err)
		}
		rerr = removeContainer(ctx, id)
		if rerr != nil {
			log.Errorln("Failed to remove container as ", rerr)
		}
	}
	log.Infoln("Action container exits with status code ", status)
	return "Successful Execution", status, nil
}

func captureLogs(ctx context.Context, id string) {
	reader, err := cli.ContainerLogs(context.Background(), id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	})
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func pullActionImage(ctx context.Context, action *pb.WorkflowAction) error {
	user := os.Getenv("REGISTRY_USERNAME")
	pwd := os.Getenv("REGISTRY_PASSWORD")
	if user == "" || pwd == "" {
		return errors.New("required REGISTRY_USERNAME and REGISTRY_PASSWORD")
	}

	authConfig := types.AuthConfig{
		Username:      user,
		Password:      pwd,
		ServerAddress: registry,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return errors.Wrap(err, "DOCKER AUTH")
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	out, err := cli.ImagePull(ctx, registry+"/"+action.GetImage(), types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		return errors.Wrap(err, "DOCKER PULL")
	}
	defer out.Close()
	if _, err := io.Copy(os.Stdout, out); err != nil {
		return err
	}
	return nil
}

func createContainer(ctx context.Context, action *pb.WorkflowAction, cmd []string, wfID string) (string, error) {
	config := &container.Config{
		Image:        registry + "/" + action.GetImage(),
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Env:          action.GetEnvironment(),
	}

	if cmd != nil {
		config.Cmd = cmd
	}

	wfDir := dataDir + string(os.PathSeparator) + wfID
	hostConfig := &container.HostConfig{
		Privileged: true,
		Binds:      []string{wfDir + ":/workflow"},
	}
	hostConfig.Binds = append(hostConfig.Binds, action.GetVolumes()...)

	log.Infoln("Starting the container with cmd", cmd)
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, action.GetName())
	if err != nil {
		return "", errors.Wrap(err, "DOCKER CREATE")
	}
	return resp.ID, nil
}

func runContainer(ctx context.Context, id string) error {
	log.Debugln("run Container with ID : ", id)
	err := cli.ContainerStart(ctx, id, types.ContainerStartOptions{})
	if err != nil {
		return errors.Wrap(err, "DOCKER START")
	}
	return nil
}

func waitContainer(ctx context.Context, id string) (pb.ActionState, error, error) {
	// Inspect whether the container is in running state
	inspect, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		log.Debugln("Container does not exists")
		return pb.ActionState_ACTION_FAILED, nil, nil
	}
	if inspect.ContainerJSONBase.State.Running {
		log.Debugln("Container with id : ", id, " is in running state")
		//return pb.ActionState_ACTION_FAILED, nil, nil
	}
	// send API call to wait for the container completion
	log.Debugln("Starting Container wait for id : ", id)
	wait, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	select {
	case status := <-wait:
		log.Infoln("Container with id ", id, "finished with status code : ", status.StatusCode)
		if status.StatusCode == 0 {
			return pb.ActionState_ACTION_SUCCESS, nil, nil
		}
		return pb.ActionState_ACTION_FAILED, nil, nil
	case err := <-errC:
		log.Errorln("Container wait failed for id : ", id, " Error : ", err)
		return pb.ActionState_ACTION_FAILED, nil, err
	case <-ctx.Done():
		log.Errorln("Container wait for id : ", id, " is timedout Error : ", err)
		return pb.ActionState_ACTION_TIMEOUT, ctx.Err(), nil
	}
}

func waitFailedContainer(ctx context.Context, id string, failedActionStatus chan pb.ActionState) {
	// send API call to wait for the container completion
	log.Debugln("Starting Container wait for id : ", id)
	wait, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	select {
	case status := <-wait:
		log.Infoln("Container with id ", id, "finished with status code : ", status.StatusCode)
		if status.StatusCode == 0 {
			failedActionStatus <- pb.ActionState_ACTION_SUCCESS
		}
		failedActionStatus <- pb.ActionState_ACTION_FAILED
	case err := <-errC:
		log.Errorln("Container wait failed for id : ", id, " Error : ", err)
		failedActionStatus <- pb.ActionState_ACTION_FAILED
	}
}

func removeContainer(ctx context.Context, id string) error {
	// create options for removing container
	opts := types.ContainerRemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: true,
	}
	log.Debugln("Start removing container ", id)
	// send API call to remove the container
	err := cli.ContainerRemove(ctx, id, opts)
	if err != nil {
		return err
	}
	return nil
}

func initializeDockerClient() (*client.Client, error) {
	registry = os.Getenv("DOCKER_REGISTRY")
	if registry == "" {
		return nil, errors.New("required DOCKER_REGISTRY")
	}
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Wrap(err, "DOCKER CLIENT")
	}
	level := os.Getenv("WORKER_LOG_LEVEL")
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
			logger.Errorln("Invalid value for WORKER_LOG_LEVEL", level, " .Setting it to default(Info)")
		}
	} else {
		logger.SetLevel(logrus.InfoLevel)
		logger.Errorln("Variable WORKER_LOG_LEVEL is not set. Default is Info")
	}
	logger.SetFormatter(&logrus.JSONFormatter{})
	return c, nil
}
