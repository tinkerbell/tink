package framework

import (
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()
var log *logrus.Entry

// Log : This Log will be used in test cases.
var Log = logger

// SetupWorkflow ... Set up workflow
func SetupWorkflow(tar string, tmpl string) (string, error) {
	//Add target machine mac/ip addr into targets table
	targetID, err := CreateTargets(tar)
	if err != nil {
		return "", err
	}
	logger.Infoln("Target Created : ", targetID)
	//Add template in template table
	templateID, err := CreateTemplate(tmpl)
	if err != nil {
		return "", err
	}
	logger.Infoln("Template Created : ", templateID)
	workflowID, err := CreateWorkflow(templateID, targetID)
	if err != nil {
		logger.Debugln("Workflow is not Created because : ", err)
		return "", err
	}
	logger.Infoln("Workflow Created : ", workflowID)
	return workflowID, nil
}
