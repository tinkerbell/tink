package executor

import (
	pb "github.com/packethost/rover/protos/rover"
)

// TODO: to be removed
func ingest() {
	workflowcontexts["1"] = &pb.WorkflowContext{
		WorkflowId: "1",
	}

	workflowactions["1"] = &pb.WorkflowActionList{
		ActionList: []*pb.WorkflowAction{
			{
				WorkerId: "worker1",
				TaskName: "task1",
				Name:     "action1",
				Image:    "task1action1",
			},
			{
				WorkerId: "worker1",
				TaskName: "task1",
				Name:     "action2",
				Image:    "task1action2",
			},
			{
				WorkerId: "worker2",
				TaskName: "task2",
				Name:     "action1",
				Image:    "task2action1",
			},
			{
				WorkerId: "worker1",
				TaskName: "task3",
				Name:     "action1",
				Image:    "task3action1",
			},
			{
				WorkerId: "worker2",
				TaskName: "task4",
				Name:     "action1",
				Image:    "task4action1",
			},
		},
	}

	for wfID, wfActions := range workflowactions {
		workflowcontexts[wfID] = &pb.WorkflowContext{
			WorkflowId: wfID,
		}
		for _, action := range wfActions.GetActionList() {
			if _, ok := workers[action.GetWorkerId()]; !ok {
				workers[action.GetWorkerId()] = []string{}
			}
			wfs := workers[action.GetWorkerId()]
			add := true
			for _, wf := range wfs {
				if wfID == wf {
					add = false
					break
				}
			}
			if add {
				workers[action.GetWorkerId()] = append(wfs, wfID)
			}
		}
	}
}
