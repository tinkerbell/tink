package workflow

import (
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	errEmptyName           = "task/action name cannot be empty"
	errInvalidLength       = "task/action name cannot have more than 200 characters: %s"
	errDuplicateTaskName   = "two tasks in a template cannot have same name: %s"
	errInvalidActionImage  = "invalid action image: %s"
	errDuplicateActionName = "two actions in a task cannot have same name: %s"
)

// Parse parses the template yaml content into a Workflow
func Parse(yamlContent []byte) (*Workflow, error) {
	var workflow Workflow

	err := yaml.UnmarshalStrict(yamlContent, &workflow)
	if err != nil {
		return &Workflow{}, errors.Wrap(err, "parsing yaml data")
	}

	if err = validate(&workflow); err != nil {
		return &Workflow{}, errors.Wrap(err, "validating workflow template")
	}

	return &workflow, nil
}

// validate validates a workflow template against certain requirements
func validate(wf *Workflow) error {
	taskNameMap := make(map[string]struct{})
	for _, task := range wf.Tasks {
		if hasEmptyName(task.Name) {
			return errors.New(errEmptyName)
		}
		if !hasValidLength(task.Name) {
			return errors.Errorf(errInvalidLength, task.Name)
		}
		_, ok := taskNameMap[task.Name]
		if ok {
			return errors.Errorf(errDuplicateTaskName, task.Name)
		}
		taskNameMap[task.Name] = struct{}{}
		actionNameMap := make(map[string]struct{})
		for _, action := range task.Actions {
			if hasEmptyName(action.Name) {
				return errors.New(errEmptyName)
			}

			if !hasValidLength(action.Name) {
				return errors.Errorf(errInvalidLength, action.Name)
			}

			if !hasValidImageName(action.Image) {
				return errors.Errorf(errInvalidActionImage, action.Image)
			}

			_, ok := actionNameMap[action.Name]
			if ok {
				return errors.Errorf(errDuplicateActionName, action.Name)
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
