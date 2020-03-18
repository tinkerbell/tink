package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	pb "github.com/packethost/tinkerbell/protos/workflow"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const workflowData = `/workflow/data:/workflow/data`

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
		return fmt.Sprintf("Failed to create container"), 1, errors.Wrap(err, "DOCKER CREATE")
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
	startedAt := time.Now()
	err = runContainer(timeCtx, id)
	if err != nil {
		return fmt.Sprintf("Failed to run container"), 1, errors.Wrap(err, "DOCKER RUN")
	}

	failedActionStatus := make(chan pb.ActionState)

	//capturing logs of action container in a go-routine
	stopLogs := make(chan bool)
	go captureLogs(ctx, startedAt, id, stopLogs)

	status, err, werr := waitContainer(timeCtx, id, stopLogs)
	if werr != nil {
		rerr := removeContainer(ctx, id)
		if rerr != nil {
			log.WithField("container_id", id).Errorln("Failed to remove container as ", rerr)
		}
		return fmt.Sprintf("Failed to wait for completion of action"), status, errors.Wrap(err, "DOCKER_WAIT")
	}
	rerr := removeContainer(ctx, id)
	if rerr != nil {
		return fmt.Sprintf("Failed to remove container of action"), status, errors.Wrap(rerr, "DOCKER_REMOVE")
	}
	log.Infoln("Container removed with Status ", pb.ActionState(status))
	if status != pb.ActionState_ACTION_SUCCESS {
		if status == pb.ActionState_ACTION_TIMEOUT && action.OnTimeout != nil {
			id, err = createContainer(ctx, action, action.OnTimeout, wfID)
			if err != nil {
				log.Errorln("Failed to create container for on-timeout command: ", err)
			}
			log.Infoln("Container created with on-timeout command : ", action.OnTimeout)
			startedAt = time.Now()
			failedActionStatus := make(chan pb.ActionState)
			go captureLogs(ctx, startedAt, id, stopLogs)
			go waitFailedContainer(ctx, id, stopLogs, failedActionStatus)
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
				go captureLogs(ctx, startedAt, id, stopLogs)
				go waitFailedContainer(ctx, id, stopLogs, failedActionStatus)
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
	return fmt.Sprintf("Successfull Execution"), status, nil
}

func captureLogs(ctx context.Context, srt time.Time, id string, stop chan bool) {
	req := func(sr string) {
		// get logs the runtime container
		rc, err := getLogs(ctx, cli, id, strconv.FormatInt(srt.Unix(), 10))
		if err != nil {
			stop <- true
		}
		defer rc.Close()
		io.Copy(os.Stdout, rc)
	}
Loop:
	for {
		select {
		case <-stop:
			// Last call to be sure to get the end of the logs content
			now := time.Now()
			now = now.Add(time.Second * -1)
			startAt := strconv.FormatInt(now.Unix(), 10)
			req(startAt)
			break Loop
		default:
			// Running call to trace the container logs every 500ms
			startAt := strconv.FormatInt(srt.Unix(), 10)
			srt = srt.Add(time.Millisecond * 500)
			req(startAt)
		}
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
	io.Copy(os.Stdout, out)
	return nil
}

func createContainer(ctx context.Context, action *pb.WorkflowAction, cmd []string, wfID string) (string, error) {
	config := &container.Config{
		Image:        registry + "/" + action.GetImage(),
		AttachStdout: true,
		AttachStderr: true,
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

func getLogs(ctx context.Context, cli *client.Client, id string, srt string) (io.ReadCloser, error) {

	log.Debugln("Capturing logs for container : ", id)
	// create options for capturing container logs
	opts := types.ContainerLogsOptions{
		Follow:     true,
		ShowStdout: true,
		ShowStderr: true,
		Details:    false,
		Since:      srt,
	}

	// send API call to capture the container logs
	logs, err := cli.ContainerLogs(ctx, id, opts)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func waitContainer(ctx context.Context, id string, stopLogs chan bool) (pb.ActionState, error, error) {
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
		stopLogs <- true
		if status.StatusCode == 0 {
			return pb.ActionState_ACTION_SUCCESS, nil, nil
		}
		return pb.ActionState_ACTION_FAILED, nil, nil
	case err := <-errC:
		log.Errorln("Container wait failed for id : ", id, " Error : ", err)
		stopLogs <- true
		return pb.ActionState_ACTION_FAILED, nil, err
	case <-ctx.Done():
		log.Errorln("Container wait for id : ", id, " is timedout Error : ", err)
		stopLogs <- true
		return pb.ActionState_ACTION_TIMEOUT, ctx.Err(), nil
	}
}

func waitFailedContainer(ctx context.Context, id string, stopLogs chan bool, failedActionStatus chan pb.ActionState) {
	// send API call to wait for the container completion
	log.Debugln("Starting Container wait for id : ", id)
	wait, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	select {
	case status := <-wait:
		log.Infoln("Container with id ", id, "finished with status code : ", status.StatusCode)
		stopLogs <- true
		if status.StatusCode == 0 {
			failedActionStatus <- pb.ActionState_ACTION_SUCCESS
		}
		failedActionStatus <- pb.ActionState_ACTION_FAILED
	case err := <-errC:
		log.Errorln("Container wait failed for id : ", id, " Error : ", err)
		stopLogs <- true
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
		return nil, errors.New("requried DOCKER_REGISTRY")
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
