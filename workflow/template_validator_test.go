package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	validTemplate = `version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "{{.device_1}}"
    actions:
    - name: "hello_world"
      image: hello-world
      timeout: 60`

	invalidTemplate = `version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "{{.device_1}}"
    actions:
  - name: "hello_world"
      image: hello-world
      timeout: 60`
)

func TestParseYAML(t *testing.T) {
	testcases := []struct {
		name          string
		content       []byte
		expectedError bool
	}{
		{
			name:    "valid template",
			content: []byte(validTemplate),
		},
		{
			name:          "invalid template",
			content:       []byte(invalidTemplate),
			expectedError: true,
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			res, err := ParseYAML([]byte(test.content))
			if err != nil {
				assert.Error(t, err)
				assert.Empty(t, res)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, res)
		})
	}
}

func TestValidateTemplate(t *testing.T) {
	testCases := []struct {
		name          string
		wf            *Workflow
		expectedError bool
	}{
		{
			name:          "invalid task name",
			wf:            workflow(withInvalidTaskName()),
			expectedError: true,
		},
		{
			name:          "long task name",
			wf:            workflow(withLongTaskName()),
			expectedError: true,
		},
		{
			name:          "duplicate task name",
			wf:            workflow(withDuplicateTaskName()),
			expectedError: true,
		},
		{
			name:          "invalid action name",
			wf:            workflow(withInvalidActionName()),
			expectedError: true,
		},
		{
			name:          "invalid action image",
			wf:            workflow(withInvalidActionImage()),
			expectedError: true,
		},
		{
			name:          "duplicate action name",
			wf:            workflow(withDuplicateActionName()),
			expectedError: true,
		},
		{
			name:          "long action name",
			wf:            workflow(withLongActionName()),
			expectedError: true,
		},
		{
			name: "valid task name",
			wf:   workflow(),
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateTemplate(test.wf)
			if err != nil && test.expectedError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

type workflowModifier func(*Workflow)

func withLongTaskName() workflowModifier {
	return func(wf *Workflow) {
		wf.Tasks[0].Name = "this task has a very long name to test whether we recevice an error or not if a task has very long name, one that would probably go beyond the limit of not having a task name with more than two hundred characters"
	}
}

func withLongActionName() workflowModifier {
	return func(wf *Workflow) {
		wf.Tasks[0].Actions[0].Name = "this action has a very long name to test whether we recevice an error or not if an action has very long name, one that would probably go beyond the limit of not having an action name with more than two hundred characters"
	}
}

func withInvalidTaskName() workflowModifier {
	return func(wf *Workflow) { wf.Tasks[0].Name = "" }
}

func withDuplicateTaskName() workflowModifier {
	return func(wf *Workflow) { wf.Tasks = append(wf.Tasks, wf.Tasks[0]) }
}

func withInvalidActionName() workflowModifier {
	return func(wf *Workflow) { wf.Tasks[0].Actions[0].Name = "" }
}

func withInvalidActionImage() workflowModifier {
	return func(wf *Workflow) { wf.Tasks[0].Actions[0].Image = "action-image-with-$#@-" }
}

func withDuplicateActionName() workflowModifier {
	return func(wf *Workflow) { wf.Tasks[0].Actions = append(wf.Tasks[0].Actions, wf.Tasks[0].Actions[0]) }
}

func workflow(m ...workflowModifier) *Workflow {
	wf := &Workflow{
		ID:            "ce2e62ed-826f-4485-a39f-a82bb74338e2",
		GlobalTimeout: 900,
		Name:          "ubuntu-provisioning",
		Tasks: []Task{
			{
				Name:       "pre-installation",
				WorkerAddr: "08:00:27:00:00:01",
				Environment: map[string]string{
					"MIRROR_HOST": "192.168.1.2",
				},
				Volumes: []string{
					"/dev:/dev",
					"/dev/console:/dev/console",
					"/lib/firmware:/lib/firmware:ro",
				},
				Actions: []Action{
					{
						Name:    "disk-wipe",
						Image:   "disk-wipe",
						Timeout: 90,
					},
					{
						Name:    "disk-partition",
						Image:   "disk-partition",
						Timeout: 300,
						Volumes: []string{
							"/statedir:/statedir",
						},
					},
					{
						Name:    "install-root-fs",
						Image:   "install-root-fs",
						Timeout: 600,
					},
					{
						Name:    "install-grub",
						Image:   "install-grub",
						Timeout: 600,
						Volumes: []string{
							"/statedir:/statedir",
						},
					},
				},
			},
		},
	}
	for _, f := range m {
		f(wf)
	}
	return wf
}
