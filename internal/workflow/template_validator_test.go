package workflow

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

const (
	validTemplate = `
version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "{{.device_1}}"
    actions:
    - name: "hello_world"
      image: hello-world
      timeout: 60
`

	invalidTemplate = `
version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "{{.device_1}}"
    actions:
  - name: "hello_world"
      image: hello-world
      timeout: 60
`

	veryLongName = "this is a very long string, that is used to test if the name is too long hahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhuhahahehehohohuhu"
)

func TestMustParse(t *testing.T) {
	table := []struct {
		Name      string
		Input     string
		WantPanic bool
	}{
		{
			Name:      "parse-valid-template",
			Input:     validTemplate,
			WantPanic: false,
		},
		{
			Name:      "parse-invalid-template",
			Input:     invalidTemplate,
			WantPanic: true,
		},
	}
	for _, s := range table {
		t.Run(s.Name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil && s.WantPanic {
					t.Errorf("panic expected but we didn't got one: %s", r)
				} else if r != nil && !s.WantPanic {
					t.Errorf("unexpected panic: %s", r)
				}
			}()

			_ = MustParse([]byte(s.Input))
		})
	}
}

func TestMustParseFromFile(t *testing.T) {
	table := []struct {
		Name      string
		Input     string
		WantPanic bool
	}{
		{
			Name:      "parse-valid-template",
			Input:     validTemplate,
			WantPanic: false,
		},
		{
			Name:      "parse-invalid-template",
			Input:     invalidTemplate,
			WantPanic: true,
		},
	}
	for _, s := range table {
		t.Run(s.Name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil && s.WantPanic {
					t.Errorf("panic expected but we didn't got one: %s", r)
				} else if r != nil && !s.WantPanic {
					t.Errorf("unexpected panic: %s", r)
				}
			}()

			file, err := os.CreateTemp(os.TempDir(), "tinktest")
			if err != nil {
				t.Error(err)
			}
			defer os.Remove(file.Name())

			err = os.WriteFile(file.Name(), []byte(s.Input), os.ModeAppend)
			if err != nil {
				t.Error(err)
			}

			_ = MustParseFromFile(file.Name())
		})
	}
}

func TestParse(t *testing.T) {
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
			res, err := Parse(test.content)
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
			name:          "template name is invalid",
			wf:            toWorkflow(withTemplateInvalidName()),
			expectedError: true,
		},
		{
			name:          "template name too long",
			wf:            toWorkflow(withTemplateLongName()),
			expectedError: true,
		},
		{
			name:          "template version is invalid",
			wf:            toWorkflow(withTemplateInvalidVersion()),
			expectedError: true,
		},
		{
			name:          "template tasks is nil",
			wf:            toWorkflow(withTemplateNilTasks()),
			expectedError: true,
		},
		{
			name:          "template tasks is empty",
			wf:            toWorkflow(withTemplateEmptyTasks()),
			expectedError: true,
		},
		{
			name:          "task name is invalid",
			wf:            toWorkflow(withTaskInvalidName()),
			expectedError: true,
		},
		{
			name:          "task name is too long",
			wf:            toWorkflow(withTaskLongName()),
			expectedError: true,
		},
		{
			name:          "task name is duplicated",
			wf:            toWorkflow(withTaskDuplicateName()),
			expectedError: true,
		},
		{
			name:          "action name is invalid",
			wf:            toWorkflow(withActionInvalidName()),
			expectedError: true,
		},
		{
			name:          "action name is duplicated",
			wf:            toWorkflow(withActionDuplicateName()),
			expectedError: true,
		},
		{
			name:          "action name is too long",
			wf:            toWorkflow(withActionLongName()),
			expectedError: true,
		},
		{
			name:          "action image is invalid",
			wf:            toWorkflow(withActionInvalidImage()),
			expectedError: true,
		},
		{
			name: "valid task name",
			wf:   toWorkflow(),
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			err := validate(test.wf)
			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name             string
		hwAddress        []byte
		templateID       string
		templateData     string
		expectedError    func(t *testing.T, err error)
		expectedTemplate string
	}{
		{
			name:         "valid-hardware-address",
			hwAddress:    []byte("{\"device_1\":\"08:00:27:00:00:01\"}"),
			templateID:   "49748301-d0d9-4ee9-84df-b64e6e1ef3dd",
			templateData: validTemplate,
			expectedTemplate: `
version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "08:00:27:00:00:01"
    actions:
    - name: "hello_world"
      image: hello-world
      timeout: 60
`,
		},
		{
			name:         "invalid-hardware-address",
			templateData: validTemplate,
			hwAddress:    []byte("{\"invalid_device\":\"08:00:27:00:00:01\"}"),
			expectedError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), `executing "workflow-template" at <.device_1>: map has no entry for key "device_1"`) {
					t.Errorf("\nexpected err: %s\ngot: %s", `executing "workflow-template" at <.device_1>: map has no entry for key "device_1"`, err)
				}
			},
		},
		{
			name:       "template with << should not be escaped in any way",
			hwAddress:  []byte("{\"device_1\":\"08:00:27:00:00:01\"}"),
			templateID: "98788301-d0d9-4ee9-84df-b64e6e1ef1cc",
			templateData: `
version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world<<"
    worker: "{{.device_1}}"
    actions:
    - name: "hello_world"
      image: hello-world
      timeout: 60
`,
			expectedTemplate: `
version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world<<"
    worker: "08:00:27:00:00:01"
    actions:
    - name: "hello_world"
      image: hello-world
      timeout: 60
`,
		},
		{
			name:       "malformed template",
			hwAddress:  []byte("{\"device_1\":\"08:00:27:00:00:01\"}"),
			templateID: "98788301-d0d9-4ee9-84df-b64e6e1ef1cc",
			templateData: `
version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "{{.device_1}"
    actions:
    - name: "hello_world"
      image: hello-world
      timeout: 60
`,
			expectedError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil {
					t.Error("expected error, got nil")
				}
				e1 := `template: workflow-template:7: unexpected "}" in operand`
				e2 := `template: workflow-template:7: bad character U+007D '}'`
				if !strings.Contains(err.Error(), e1) && !strings.Contains(err.Error(), e2) {
					base := "failed to parse template with ID 98788301-d0d9-4ee9-84df-b64e6e1ef1cc: "
					t.Errorf("\nexpected err:    '%s'\nor expected err: '%s'\ngot:             '%s'", base+e1, base+e2, err)
				}
			},
		},
		{
			name:       "invalid yaml in template",
			hwAddress:  []byte("{\"device_1\":\"08:00:27:00:00:01\"}"),
			templateID: "98788301-d0d9-4ee9-84df-b64e6e1ef1cc",
			templateData: `
version: "0.1"
  name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "{{.device_1}}"
    actions:
    - name: "hello_world"
      image: hello-world
      timeout: 60
`,
			expectedError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), `parsing yaml data: yaml: line 2: did not find expected key`) {
					t.Errorf("\nexpected err: '%s'\ngot:          '%s'", `parsing yaml data: yaml: line 2: did not find expected key`, err)
				}
			},
		},
		{
			name:       "Empty worker address",
			hwAddress:  []byte("{\"device_1\":\"\"}"),
			templateID: "98788301-d0d9-4ee9-84df-b64e6e1ef1cc",
			templateData: `
version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "{{.device_1}}"
    actions:
    - name: "hello_world"
      image: hello-world
      timeout: 60
`,
			expectedError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), `failed to render template, invalid hardware address: map[device_1:]`) {
					t.Errorf("\nexpected err: '%s'\ngot:          '%s'", `failed to render template, invalid hardware address: map[device_1:]`, err)
				}
			},
		},
		{
			name:       "Invalid hardware json",
			hwAddress:  []byte(`{"device_1":"abc"}}`),
			templateID: "98788301-d0d9-4ee9-84df-b64e6e1ef1cc",
			templateData: `
version: "0.1"
name: hello_world_workflow
global_timeout: 600
tasks:
  - name: "hello world"
    worker: "{{.device_1}}"
    actions:
    - name: "hello_world"
      image: hello-world
      timeout: 60
`,
			expectedError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !strings.Contains(err.Error(), `failed to parse template with ID 98788301-d0d9-4ee9-84df-b64e6e1ef1cc: invalid character '}' after top-level value`) {
					t.Errorf("\nexpected err: '%s'\ngot:          '%s'", `failed to parse template with ID 98788301-d0d9-4ee9-84df-b64e6e1ef1cc: invalid character '}' after top-level value`, err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			temp, err := RenderTemplate(test.templateID, test.templateData, test.hwAddress)
			if test.expectedError != nil {
				test.expectedError(t, err)
				return
			}
			if diff := cmp.Diff(test.expectedTemplate, temp); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRenderTemplateHardwareCustomFuncs(t *testing.T) {
	cases := []struct {
		name         string
		hwAddress    []byte
		templateID   string
		templateData string
		expected     Workflow
	}{
		{
			name:       "contains/isFalse",
			templateID: "test",
			templateData: `
version: "0.1"
name: test
global_timeout: 1
tasks:
  - name: "test{{ if ( contains "foo" "bar" ) }}test{{ end }}"
    worker: "test"
    actions:
    - name: "test"
      image: test
      timeout: 60
`,
			expected: Workflow{
				Name:          "test",
				GlobalTimeout: 1,
				Version:       "0.1",
				Tasks: []Task{
					{
						Name:       "test",
						WorkerAddr: "test",
						Actions: []Action{
							{
								Name:    "test",
								Image:   "test",
								Timeout: 60,
							},
						},
					},
				},
			},
		},
		{
			name:       "contains/isTrue",
			templateID: "test",
			templateData: `
version: "0.1"
name: test
global_timeout: 1
tasks:
  - name: "test{{ if ( contains "foo" "foo" ) }}test{{ end }}"
    worker: "test"
    actions:
    - name: "test"
      image: test
      timeout: 60
`,
			expected: Workflow{
				Name:          "test",
				GlobalTimeout: 1,
				Version:       "0.1",
				Tasks: []Task{
					{
						Name:       "testtest",
						WorkerAddr: "test",
						Actions: []Action{
							{
								Name:    "test",
								Image:   "test",
								Timeout: 60,
							},
						},
					},
				},
			},
		},
		{
			name:       "hasPrefix/isFalse",
			templateID: "test",
			templateData: `
version: "0.1"
name: test
global_timeout: 1
tasks:
  - name: "test{{ if ( hasPrefix "foo" "bar" ) }}test{{ end }}"
    worker: "test"
    actions:
    - name: "test"
      image: test
      timeout: 60
`,
			expected: Workflow{
				Name:          "test",
				GlobalTimeout: 1,
				Version:       "0.1",
				Tasks: []Task{
					{
						Name:       "test",
						WorkerAddr: "test",
						Actions: []Action{
							{
								Name:    "test",
								Image:   "test",
								Timeout: 60,
							},
						},
					},
				},
			},
		},
		{
			name:       "hasPrefix/isTrue",
			templateID: "test",
			templateData: `
version: "0.1"
name: test
global_timeout: 1
tasks:
  - name: "test{{ if ( hasPrefix "foo" "foo" ) }}test{{ end }}"
    worker: "test"
    actions:
    - name: "test"
      image: test
      timeout: 60
`,
			expected: Workflow{
				Name:          "test",
				GlobalTimeout: 1,
				Version:       "0.1",
				Tasks: []Task{
					{
						Name:       "testtest",
						WorkerAddr: "test",
						Actions: []Action{
							{
								Name:    "test",
								Image:   "test",
								Timeout: 60,
							},
						},
					},
				},
			},
		},
		{
			name:       "hasSuffix/isFalse",
			templateID: "test",
			templateData: `
version: "0.1"
name: test
global_timeout: 1
tasks:
  - name: "test{{ if ( hasSuffix "foo" "bar" ) }}test{{ end }}"
    worker: "test"
    actions:
    - name: "test"
      image: test
      timeout: 60
`,
			expected: Workflow{
				Name:          "test",
				GlobalTimeout: 1,
				Version:       "0.1",
				Tasks: []Task{
					{
						Name:       "test",
						WorkerAddr: "test",
						Actions: []Action{
							{
								Name:    "test",
								Image:   "test",
								Timeout: 60,
							},
						},
					},
				},
			},
		},
		{
			name:       "hasSuffix/isTrue",
			templateID: "test",
			templateData: `
version: "0.1"
name: test
global_timeout: 1
tasks:
  - name: "test{{ if ( hasSuffix "foo" "foo" ) }}test{{ end }}"
    worker: "test"
    actions:
    - name: "test"
      image: test
      timeout: 60
`,
			expected: Workflow{
				Name:          "test",
				GlobalTimeout: 1,
				Version:       "0.1",
				Tasks: []Task{
					{
						Name:       "testtest",
						WorkerAddr: "test",
						Actions: []Action{
							{
								Name:    "test",
								Image:   "test",
								Timeout: 60,
							},
						},
					},
				},
			},
		},
		{
			name:       "formatPartition/block device",
			templateID: "test",
			templateData: `
version: "0.1"
name: test
global_timeout: 1
tasks:
  - name: "test"
    worker: "test"
    actions:
    - name: "test"
      image: test
      timeout: 60
      environment:
        DEST_DISK: {{ formatPartition "/dev/sda" 1 }}
`,
			expected: Workflow{
				Name:          "test",
				GlobalTimeout: 1,
				Version:       "0.1",
				Tasks: []Task{
					{
						Name:       "test",
						WorkerAddr: "test",
						Actions: []Action{
							{
								Name:    "test",
								Image:   "test",
								Timeout: 60,
								Environment: map[string]string{
									"DEST_DISK": "/dev/sda1",
								},
							},
						},
					},
				},
			},
		},
		{
			name:       "formatPartition/nvme device",
			templateID: "test",
			templateData: `
version: "0.1"
name: test
global_timeout: 1
tasks:
  - name: "test"
    worker: "test"
    actions:
    - name: "test"
      image: test
      timeout: 60
      environment:
        DEST_DISK: {{ formatPartition "/dev/nvme0n1" 1 }}
`,
			expected: Workflow{
				Name:          "test",
				GlobalTimeout: 1,
				Version:       "0.1",
				Tasks: []Task{
					{
						Name:       "test",
						WorkerAddr: "test",
						Actions: []Action{
							{
								Name:    "test",
								Image:   "test",
								Timeout: 60,
								Environment: map[string]string{
									"DEST_DISK": "/dev/nvme0n1p1",
								},
							},
						},
					},
				},
			},
		},
		{
			name:       "formatPartition/unknown",
			templateID: "test",
			templateData: `
version: "0.1"
name: test
global_timeout: 1
tasks:
  - name: "test"
    worker: "test"
    actions:
    - name: "test"
      image: test
      timeout: 60
      environment:
        DEST_DISK: {{ formatPartition "/dev/foobar" 1 }}
`,
			expected: Workflow{
				Name:          "test",
				GlobalTimeout: 1,
				Version:       "0.1",
				Tasks: []Task{
					{
						Name:       "test",
						WorkerAddr: "test",
						Actions: []Action{
							{
								Name:    "test",
								Image:   "test",
								Timeout: 60,
								Environment: map[string]string{
									"DEST_DISK": "/dev/foobar",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wflw, _, err := RenderTemplateHardware(tc.templateID, tc.templateData, map[string]interface{}{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(*wflw, tc.expected); diff != "" {
				t.Errorf("unexpected workflow (-want +got):\n%s", diff)
			}
		})
	}
}

type workflowModifier func(*Workflow)

func toWorkflow(m ...workflowModifier) *Workflow {
	wf := &Workflow{
		ID:            "ce2e62ed-826f-4485-a39f-a82bb74338e2",
		GlobalTimeout: 900,
		Name:          "ubuntu-provisioning",
		Version:       "0.1",
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

// invalid task modifiers

func withTaskInvalidName() workflowModifier {
	return func(wf *Workflow) { wf.Tasks[0].Name = "" }
}

func withTaskLongName() workflowModifier {
	return func(wf *Workflow) {
		wf.Tasks[0].Name = veryLongName
	}
}

func withTaskDuplicateName() workflowModifier {
	return func(wf *Workflow) { wf.Tasks = append(wf.Tasks, wf.Tasks[0]) }
}

// invalid action modifiers

func withActionInvalidName() workflowModifier {
	return func(wf *Workflow) { wf.Tasks[0].Actions[0].Name = "" }
}

func withActionLongName() workflowModifier {
	return func(wf *Workflow) {
		wf.Tasks[0].Actions[0].Name = veryLongName
	}
}

func withActionDuplicateName() workflowModifier {
	return func(wf *Workflow) { wf.Tasks[0].Actions = append(wf.Tasks[0].Actions, wf.Tasks[0].Actions[0]) }
}

func withActionInvalidImage() workflowModifier {
	return func(wf *Workflow) { wf.Tasks[0].Actions[0].Image = "action-image-with-$#@-" }
}

// invalid template modifiers

func withTemplateInvalidName() workflowModifier {
	return func(wf *Workflow) { wf.Name = "" }
}

func withTemplateLongName() workflowModifier {
	return func(wf *Workflow) {
		wf.Name = veryLongName
	}
}

func withTemplateInvalidVersion() workflowModifier {
	return func(wf *Workflow) {
		wf.Version = "0.2"
	}
}

func withTemplateNilTasks() workflowModifier {
	return func(wf *Workflow) {
		wf.Tasks = nil
	}
}

func withTemplateEmptyTasks() workflowModifier {
	return func(wf *Workflow) {
		wf.Tasks = []Task{}
	}
}
