package framework

import (
	"context"

	"github.com/packethost/rover/client"
	"github.com/packethost/rover/protos/workflow"
)

func CreateWorkflow(template string, target string) (string, error) {
	req := workflow.CreateRequest{Template: template, Target: target}
	res, err := client.WorkflowClient.CreateWorkflow(context.Background(), &req)
	if err != nil {
		return "", err
	}
	return res.Id, nil
}
