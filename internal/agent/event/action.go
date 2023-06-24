package event

import "fmt"

const (
	ActionStartedName   Name = "ActionStarted"
	ActionSucceededName Name = "ActionSucceeded"
	ActionFailedName    Name = "ActionFailed"
)

// ActionStarted occurs when an action begins running.
type ActionStarted struct {
	ActionID   string
	WorkflowID string
}

func (ActionStarted) GetName() Name {
	return ActionStartedName
}

func (e ActionStarted) String() string {
	return fmt.Sprintf("workflow=%v action=%v", e.WorkflowID, e.ActionID)
}

// ActionSucceeded occurs when an action successfully completes.
type ActionSucceeded struct {
	ActionID   string
	WorkflowID string
}

func (ActionSucceeded) GetName() Name {
	return ActionSucceededName
}

func (e ActionSucceeded) String() string {
	return fmt.Sprintf("workflow=%v action=%v", e.WorkflowID, e.ActionID)
}

// ActionFailed occurs when an action fails to complete.
type ActionFailed struct {
	ActionID   string
	WorkflowID string
	Reason     string
	Message    string
}

func (ActionFailed) GetName() Name {
	return ActionFailedName
}

func (e ActionFailed) String() string {
	return fmt.Sprintf("workflow='%v' action='%v' reason='%v'", e.WorkflowID, e.ActionID, e.Reason)
}
