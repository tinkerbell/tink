// Package db - workflow tests
// The following tests validate database functionality, the instance payload.
// A postgres instance with the tink database schema is required in order to run
// the following tests.
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	pb "github.com/tinkerbell/tink/protos/workflow"
)

// TestCreateWorkflow : validates the creation of a new workflow.
func TestCreateWorkflow(t *testing.T) {
	TestCreateTemplate(t)
	setupInsertIntoDB(t)
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	id, err := uuid.Parse(ids.templateID)
	if err != nil {
		t.Fail()
	}
	if _, err := a.tinkDb.instance.Exec(`delete from workflow;`); err != nil {
		t.Fail()
	}
	if _, err := a.tinkDb.instance.Exec(`delete from workflow_event;`); err != nil {
		t.Fail()
	}
	if _, err := a.tinkDb.instance.Exec(`delete from workflow_state;`); err != nil {
		t.Fail()
	}
	if _, err := a.tinkDb.instance.Exec(`delete from workflow_data;`); err != nil {
		t.Fail()
	}
	if err := a.tinkDb.CreateWorkflow(a.ctx, a.workflowData, a.templateData, id); err != nil {
		t.Logf("WORKFLOW: %+v \n", a.workflowData)
		a.logger.Error(err)
		t.Fail()
	}
}

// TestInsertIntoWfDataTable : validates the insert operation.
func TestInsertIntoWfDataTable(t *testing.T) {
	TestCreateWorkflow(t)
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	data, err := json.Marshal(a.workflowData)
	if err != nil {
		t.Fail()
	}
	update := &pb.UpdateWorkflowDataRequest{Metadata: data, Data: data, WorkflowId: ids.templateID}
	if a.tinkDb.instance, err = sql.Open("postgres", a.psqlInfo); err != nil {
		t.Fail()
	}
	if err := a.tinkDb.InsertIntoWfDataTable(a.ctx, update); err != nil {
		a.logger.Error(err)
		t.Fail()
	}
}

// TestGetfromWfDataTable : validates returning workflow data.
func TestGetfromWfDataTable(t *testing.T) {
	TestCreateWorkflow(t)
	var response []byte
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	req := &pb.GetWorkflowDataRequest{WorkflowId: ids.templateID}
	if a.tinkDb.instance, err = sql.Open("postgres", a.psqlInfo); err != nil {
		t.Fail()
	}
	if response, err = a.tinkDb.GetfromWfDataTable(a.ctx, req); err != nil {
		a.logger.Error(err)
		t.Error(err)
	}
	t.Logf("%s", response)
}

// TestGetWorkflowMetadata : validates retrieving workflow meta data.
func TestGetWorkflowMetadata(t *testing.T) {
	TestCreateWorkflow(t)
	var response []byte
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	req := &pb.GetWorkflowDataRequest{WorkflowId: ids.templateID}
	if a.tinkDb.instance, err = sql.Open("postgres", a.psqlInfo); err != nil {
		t.Fail()
	}
	if response, err = a.tinkDb.GetWorkflowMetadata(a.ctx, req); err != nil {
		a.logger.Error(err)
		t.Error(err)
	}
	t.Logf("%s", response)
}

// TestGetWorkflowDataVersion : validates getting workflow data version.
func TestGetWorkflowDataVersion(t *testing.T) {
	TestInsertIntoWfDataTable(t)
	var response int32
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if response, err = a.tinkDb.GetWorkflowDataVersion(a.ctx, ids.templateID); err != nil {
		a.logger.Error(err)
		t.Error(err)
	}
	expected := int32(1)
	if expected != response {
		t.Error(fmt.Errorf("EXPECTED: %v\nRECEIVED: %v", expected, response))
	}
}

// TestGetWorkflowsForWorker : validates getting workflows for worker.
func TestGetWorkflowsForWorker(t *testing.T) {
	TestCreateWorkflow(t)
	var response []string
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if response, err = a.tinkDb.GetWorkflowsForWorker(ids.workflowID); err != nil {
		a.logger.Error(err)
		t.Error(err)
	}
	t.Logf("%+v", response)
}

// TestGetWorkflow : validates getting workflow.
func TestGetWorkflow(t *testing.T) {
	TestCreateWorkflow(t)
	var response Workflow
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if response, err = a.tinkDb.GetWorkflow(a.ctx, ids.workflowID); err != nil {
		a.logger.Error(err)
		t.Error(err)
	}
	if a.workflowData.ID != response.ID {
		t.Errorf("EXPECTED: %v\nRECEIVED:%v", a.workflowData.ID, response.ID)
	}
}

// TestDeleteWorkflow : validates deleting workflow.
func TestDeleteWorkflow(t *testing.T) {
	TestCreateWorkflow(t)
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if a.tinkDb.instance, err = sql.Open("postgres", a.psqlInfo); err != nil {
		t.Fail()
	}
	if err = a.tinkDb.DeleteWorkflow(a.ctx, ids.workflowID, int32(0)); err != nil {
		a.logger.Error(err)
		t.Error(err)
	}
}

// TestUpdateWorkflow : validates updating a workflow.
func TestUpdateWorkflow(t *testing.T) {
	TestCreateWorkflow(t)
	var response Workflow
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if a.tinkDb.instance, err = sql.Open("postgres", a.psqlInfo); err != nil {
		t.Fail()
	}
	a.workflowData.Hardware = `{"id": "` + ids.hardwareID + `", "network": {"interfaces": [{"dhcp": {"ip": {"address": "192.168.1.6", "gateway": "192.168.1.1", "netmask": "255.255.255.248"}, "mac": "08:00:27:00:00:01", "arch": "x86_64", "uefi": false}, "netboot": {"allow_pxe": true, "allow_workflow": true}}]}, "metadata": {"state": "", "facility": {"facility_code": "onprem"}, "instance": {"ip_addresses": [{"address": ""}]}}}`
	if err = a.tinkDb.UpdateWorkflow(a.ctx, a.workflowData, int32(0)); err != nil {
		a.logger.Error(err)
		t.Error(err)
	}
	if response, err = a.tinkDb.GetWorkflow(a.ctx, ids.workflowID); err != nil {
		a.logger.Error(err)
		t.Error(err)
	}
	if response.Hardware != a.workflowData.Hardware {
		t.Error(fmt.Errorf("EXPECTED: %s\nRECEIVED:%s", a.workflowData.Hardware, response.Hardware))
	}
}

// TestUpdateWorkflowState : validates updating workflow state.
func TestUpdateWorkflowState(t *testing.T) {
	TestCreateWorkflow(t)
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if a.tinkDb.instance, err = sql.Open("postgres", a.psqlInfo); err != nil {
		t.Fail()
	}
	wfContext := &pb.WorkflowContext{WorkflowId: ids.templateID}
	if err = a.tinkDb.UpdateWorkflowState(a.ctx, wfContext); err != nil {
		a.logger.Error(err)
		t.Error(err)
	}
}

// TestGetWorkflowContexts : validates getting workflow context.
func TestGetWorkflowContexts(t *testing.T) {
	TestCreateWorkflow(t)
	var response *pb.WorkflowContext
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if a.tinkDb.instance, err = sql.Open("postgres", a.psqlInfo); err != nil {
		t.Fail()
	}
	if response, err = a.tinkDb.GetWorkflowContexts(a.ctx, ids.templateID); err != nil {
		a.logger.Error(err)
		t.Error(err)
	}
	t.Logf("%s", response)
}
