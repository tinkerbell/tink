package framework

import (
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()
var log *logrus.Entry

// Log : This Log will be used in test cases.
var Log = logger

// SetupWorkflow ... Set up workflow
func SetupWorkflow(hMAC string, tmpl string) (string, error) {
	//hardwareID := `{"device_1":"98:03:9b:89:d7:ba"}`
	//Add template in template table
	templateID, err := CreateTemplate(tmpl)
	if err != nil {
		return "", err
	}
	logger.Infoln("Template Created : ", templateID)
	hardwareID := `{"device_1":"` + hMAC + `"}`
	workflowID, err := CreateWorkflow(templateID, hardwareID)
	if err != nil {
		logger.Debugln("Workflow is not Created because : ", err)
		return "", err
	}
	logger.Infoln("Workflow Created : ", workflowID)
	return workflowID, nil
}
