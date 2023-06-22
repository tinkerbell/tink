// Package workflow contains workflow domain objects. The domain objects will be moved to
// /internal/workflow at a later date when they are required/we transition to the new codebase.
package workflow

// Workflow represents a runnable workflow for the Handler.
type Workflow struct {
	// Do we need a workflow name? Does that even come down in the proto definition?
	ID      string   `yaml:"id"`
	Actions []Action `yaml:"actions"`
}

func (w Workflow) String() string {
	return w.ID
}

// Action represents an individually runnable action.
type Action struct {
	ID               string            `yaml:"id"`
	Name             string            `yaml:"name"`
	Image            string            `yaml:"image"`
	Cmd              string            `yaml:"cmd"`
	Args             []string          `yaml:"args"`
	Env              map[string]string `yaml:"env"`
	Volumes          []string          `yaml:"volumes"`
	NetworkNamespace string            `yaml:"networkNamespace"`
}

func (a Action) String() string {
	// We should consider normalizing the action name and combining it with the ID. It would
	// make human identification easier. Alternatively, we could have a dedicated method for
	// retrieving names.
	return a.ID
}
