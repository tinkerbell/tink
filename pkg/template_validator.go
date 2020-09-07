package pkg

import (
	"fmt"

	"github.com/docker/distribution/reference"
	"gopkg.in/yaml.v2"
)

const (
	errEmptyName           = "task/action name cannot be empty: %v"
	errInvalidLength       = "task/action name cannot have more than 200 characters: %v"
	errDuplicateTaskName   = "two tasks in a template cannot have same name: %v"
	errInvalidActionImage  = "invalid action image: %v"
	errDuplicateActionName = "two actions in a task cannot have same name: %v"
	errNoTimeout           = "action %v has no timeout field or 0 timeout"
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

			err = hasValidTimeout(action.Timeout, action.Name)
			if err != nil {
				return err
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
	if name == "" {
		return fmt.Errorf(errEmptyName, name)
	}
	if len(name) > 200 {
		return fmt.Errorf(errInvalidLength, name)
	}
	return nil
}

func isValidImageName(name string) error {
	_, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return err
	}
	return nil
}

func hasValidTimeout(timeout int64, name string) error {
	if timeout > 0 {
		return nil
	}
	return fmt.Errorf(errNoTimeout, name)
}
