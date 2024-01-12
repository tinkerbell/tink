package workflow

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/distribution/reference"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	errInvalidLength   = "name cannot be empty or have more than 200 characters: %s"
	errTemplateParsing = "failed to parse template with ID %s"
)

// parse parses the template yaml content into a Workflow.
func parse(yamlContent []byte) (*Workflow, error) {
	var workflow Workflow

	if err := yaml.UnmarshalStrict(yamlContent, &workflow); err != nil {
		return &Workflow{}, errors.Wrap(err, "parsing yaml data")
	}

	if err := validate(&workflow); err != nil {
		return &Workflow{}, errors.Wrap(err, "validating workflow template")
	}

	return &workflow, nil
}

// renderTemplateHardware renders the workflow template and returns the Workflow and the interpolated bytes.
func renderTemplateHardware(templateID, templateData string, hardware map[string]interface{}) (*Workflow, error) {
	t := template.New("workflow-template").
		Option("missingkey=error").
		Funcs(templateFuncs)

	_, err := t.Parse(templateData)
	if err != nil {
		err = errors.Wrapf(err, errTemplateParsing, templateID)
		return nil, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, hardware); err != nil {
		err = errors.Wrapf(err, errTemplateParsing, templateID)
		return nil, err
	}

	wf, err := parse(buf.Bytes())
	if err != nil {
		return nil, err
	}

	for _, task := range wf.Tasks {
		if task.WorkerAddr == "" {
			return nil, fmt.Errorf("failed to render template, empty hardware address (%v)", hardware)
		}
	}

	return wf, nil
}

// validate validates a workflow template against certain requirements.
func validate(wf *Workflow) error {
	if !hasValidLength(wf.Name) {
		return errors.Errorf(errInvalidLength, wf.Name)
	}

	if wf.Version != "0.1" {
		return errors.Errorf("invalid template version (%s)", wf.Version)
	}

	if len(wf.Tasks) == 0 {
		return errors.New("template must have at least one task defined")
	}

	taskNameMap := make(map[string]struct{})
	for _, task := range wf.Tasks {
		if !hasValidLength(task.Name) {
			return errors.Errorf(errInvalidLength, task.Name)
		}

		if _, ok := taskNameMap[task.Name]; ok {
			return errors.Errorf("two tasks in a template cannot have same name (%s)", task.Name)
		}

		taskNameMap[task.Name] = struct{}{}
		actionNameMap := make(map[string]struct{})
		for _, action := range task.Actions {
			if !hasValidLength(action.Name) {
				return errors.Errorf(errInvalidLength, action.Name)
			}

			if err := validateImageName(action.Image); err != nil {
				return errors.Errorf("invalid action image (%s): %v", action.Image, err)
			}

			_, ok := actionNameMap[action.Name]
			if ok {
				return errors.Errorf("two actions in a task cannot have same name: %s", action.Name)
			}
			actionNameMap[action.Name] = struct{}{}
		}
	}
	return nil
}

func hasValidLength(name string) bool {
	return len(name) > 0 && len(name) < 200
}

func validateImageName(name string) error {
	_, err := reference.ParseNormalizedNamed(name)
	return err
}
