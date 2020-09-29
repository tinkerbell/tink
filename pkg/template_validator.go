package pkg

import (
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	errEmptyName           = "task/action name cannot be empty: "
	errInvalidLength       = "task/action name cannot have more than 200 characters: "
	errDuplicateTaskName   = "two tasks in a template cannot have same name: "
	errInvalidActionImage  = "invalid action image: "
	errDuplicateActionName = "two actions in a task cannot have same name: "
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
		if hasEmptyName(task.Name) {
			return errors.New(errEmptyName + task.Name)
		}
		if !hasValidLength(task.Name) {
			return errors.New(errInvalidLength + task.Name)
		}
		_, ok := taskNameMap[task.Name]
		if ok {
			return errors.New(errDuplicateTaskName + task.Name)
		}
		taskNameMap[task.Name] = struct{}{}
		actionNameMap := make(map[string]struct{})
		for _, action := range task.Actions {
			if hasEmptyName(action.Name) {
				return errors.New(errEmptyName + action.Name)
			}

			if !hasValidLength(action.Name) {
				return errors.New(errInvalidLength + action.Name)
			}

			if !hasValidImageName(action.Image) {
				return errors.New(errInvalidActionImage + action.Image)
			}

			_, ok := actionNameMap[action.Name]
			if ok {
				return errors.New(errDuplicateActionName + action.Name)
			}
			actionNameMap[action.Name] = struct{}{}
		}
	}
	return nil
}

func hasEmptyName(name string) bool {
	return name == ""
}
func hasValidLength(name string) bool {
	return len(name) < 200
}

func hasValidImageName(name string) bool {
	_, err := reference.ParseNormalizedNamed(name)
	return err == nil
}
