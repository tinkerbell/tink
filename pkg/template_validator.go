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
		err := hasValidLength(task.Name)
		if err != nil {
			return err
		}
		_, ok := taskNameMap[task.Name]
		if ok {
			return fmt.Errorf(errDuplicateTaskName, task.Name)
		}
		taskNameMap[task.Name] = struct{}{}
		actionNameMap := make(map[string]struct{})
		for _, action := range task.Actions {
			err := hasValidLength(action.Name)
			if err != nil {
				return err
			}
			err = isValidImageName(action.Image)
			if err != nil {
				return fmt.Errorf(errInvalidActionImage, action.Image)
			}

			_, ok := actionNameMap[action.Name]
			if ok {
				return fmt.Errorf(errDuplicateActionName, action.Name)
			}
			actionNameMap[action.Name] = struct{}{}
		}
	}
	return nil
}

func hasValidLength(name string) error {
	if len(name) > 200 {
		return fmt.Errorf(errInvalidLength, name)
	}
	return nil
}

func isValidImageName(name string) error {
	_, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
