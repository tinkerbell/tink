package framework

import "fmt"

// SetupWorkflow ... Set up workflow
func SetupWorkflow(tar string, tmpl string) (string, error) {

	//Add target machine mac/ip addr into targets table
	targetID, err := CreateTargets(tar)
	if err != nil {
		return "", err
	}
	fmt.Println("Target Created : ", targetID)
	//Add template in template table
	templateID, err := CreateTemplate(tmpl)
	if err != nil {
		return "", err
	}
	fmt.Println("Template Created : ", templateID)
	workflowID, err := CreateWorkflow(templateID, targetID)
	if err != nil {
		return "", err
	}
	fmt.Println("Workflow Created : ", workflowID)
	return workflowID, nil
}
