package workflow

import "time"

// Workflow represents a workflow to be executed
type Workflow struct {
	Version       string `yaml:"version"`
	Name          string `yaml:"name"`
	ID            string `yaml:"id"`
	GlobalTimeout int    `yaml:"global_timeout"`
	Tasks         []Task `yaml:"tasks"`
}

// Task represents a task to be executed as part of a workflow
type Task struct {
	Name        string            `yaml:"name"`
	WorkerAddr  string            `yaml:"worker"`
	Actions     []Action          `yaml:"actions"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

// Action is the basic executional unit for a workflow
type Action struct {
	Name        string            `yaml:"name"`
	Image       string            `yaml:"image"`
	Timeout     int64             `yaml:"timeout"`
	Command     []string          `yaml:"command,omitempty"`
	OnTimeout   []string          `yaml:"on-timeout,omitempty"`
	OnFailure   []string          `yaml:"on-failure,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

// Template represents a workflow template
type Template struct {
	ID        string    `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	DeletedAt time.Time `json:"deleted_at,omitempty"`
	Data      string    `json:"data,omitempty"`
}
