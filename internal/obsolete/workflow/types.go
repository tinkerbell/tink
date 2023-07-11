package workflow

// Workflow represents a workflow to be executed.
type Workflow struct {
	Version       string `yaml:"version"`
	Name          string `yaml:"name"`
	ID            string `yaml:"id"`
	GlobalTimeout int    `yaml:"global_timeout"`
	Tasks         []Task `yaml:"tasks"`
}

// Task represents a task to be executed as part of a workflow.
type Task struct {
	Name        string            `yaml:"name"`
	WorkerAddr  string            `yaml:"worker"`
	Actions     []Action          `yaml:"actions"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

// Action is the basic executional unit for a workflow.
type Action struct {
	Name        string            `yaml:"name"`
	Image       string            `yaml:"image"`
	Timeout     int64             `yaml:"timeout"`
	Command     []string          `yaml:"command,omitempty"`
	OnTimeout   []string          `yaml:"on-timeout,omitempty"`
	OnFailure   []string          `yaml:"on-failure,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Pid         string            `yaml:"pid,omitempty"`
}
