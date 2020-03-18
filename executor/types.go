package executor

// Workflow represents a workflow to be executed
type Workflow struct {
	Version       string `yaml:"version"`
	Name          string `yaml:"name"`
	GlobalTimeout int    `yaml:"global_timeout"`
	Tasks         []Task `yaml:"tasks"`
}

// Task represents a task to be executed as part of a workflow
type Task struct {
	Name    string   `yaml:"name"`
	Worker  string   `yaml:"worker"`
	Actions []Action `yaml:"actions"`
}

// Action represents an action to be executed as part of a task
type Action struct {
	Name      string `yaml:"name"`
	Image     string `yaml:"image"`
	Timeout   int    `yaml:"timeout"`
	Command   string `yaml:"command"`
	OnTimeout string `yaml:"on-timeout"`
	OnFailure string `yaml:"on-failure"`
}
