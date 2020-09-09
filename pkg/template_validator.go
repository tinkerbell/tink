package pkg

import (
	"fmt"

	"github.com/docker/distribution/reference"
	"gopkg.in/yaml.v2"
)

const (
	errInvalidLength       = "task/action name cannot have more than 200 characters: %v"
	errDuplicateTaskName   = "two tasks in a template cannot have same name: %v"
	errInvalidActionImage  = "invalid action image: %v"
	errDuplicateActionName = "two actions in a task cannot have same name: %v"
)

// ParseYAML parses the template yaml content
func ParseYAML(yamlContent []byte) (*Workflow, error) {
	var workflow = Workflow{}
	err := yaml.UnmarshalStrict(yamlContent, &workflow)
	if err != nil {
		return &Workflow{}, err
	}
	return &workflow, nil
}

// ValidateTemplate validates a workflow template
// against certain design paradigms
func ValidateTemplate(wf *Workflow) error {
	taskNameMap := make(map[string]struct{})
	for _, task := range wf.Tasks {
		if !hasValidLength(task.Name) {
			return fmt.Errorf(errInvalidLength, task.Name)
		}
		if _, taskAlreadyExists := taskNameMap[task.Name]; taskAlreadyExists {
			return fmt.Errorf(errDuplicateTaskName, task.Name)
		}
		taskNameMap[task.Name] = struct{}{}
		actionNameMap := make(map[string]struct{})
		for _, action := range task.Actions {
			if !hasValidLength(action.Name) {
				return fmt.Errorf(errInvalidLength, action.Name)
			}
			if !hasValidImageName(action.Image) {
				return fmt.Errorf(errInvalidActionImage, action.Image)
			}
			if _, actionAlreadyExists := actionNameMap[action.Name]; actionAlreadyExists {
				return fmt.Errorf(errDuplicateActionName, action.Name)
			}
			actionNameMap[action.Name] = struct{}{}
		}
	}
	return nil
}

func hasValidLength(name string) bool {
	return name != "" && len(name) <= 200
}

func hasValidImageName(name string) bool {
	_, err := reference.ParseNormalizedNamed(name)
	return err == nil
}
