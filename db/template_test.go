// Package db - template tests
// The following tests validate database functionality, the instance payload.
// A postgres instance with the tink database schema is required in order to run
// the following tests.
package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// TestCreateTemplate : validates the creation of a new workflow template.
func TestCreateTemplate(t *testing.T) {
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	id, err := uuid.Parse(ids.templateID)
	if err != nil {
		t.Fail()
	}
	if _, err := a.tinkDb.instance.Exec("delete from template;"); err != nil {
		t.Fail()
	}
	if err := a.tinkDb.CreateTemplate(a.ctx, "foo", a.templateData, id); err != nil {
		a.logger.Error(err)
		t.Fail()
	}
}

// TestCreateDuplicateTemplate : creates a duplicate template with the same name.
// This test should fail. It should only work if the original record is marked
// for deletion.
func TestCreateDuplicateTemplate(t *testing.T) {
	TestCreateTemplate(t)
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	id, err := uuid.NewUUID()
	if err != nil {
		t.Fail()
	}
	err = a.tinkDb.CreateTemplate(a.ctx, "foo", a.templateData, id)
	if err != nil {
		a.logger.Error(err)
		if !strings.Contains(err.Error(), "exists") {
			t.Error("function should return: foo already exists")
		}
	} else {
		t.Error("function should return an error")
	}
}

// TestCreateNewDuplicateTemplate : creates a duplicate template with the same
// name, but the previous template will be marked for deletion.
// This test should pass.
func TestCreateNewDuplicateTemplate(t *testing.T) {
	setupInsertIntoDB(t)
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if err = a.tinkDb.DeleteTemplate(a.ctx, ids.templateID); err != nil {
		a.logger.Error(err)
		t.Fail()
	}
	id, err := uuid.NewUUID()
	if err != nil {
		t.Fail()
	}
	if err = a.tinkDb.CreateTemplate(a.ctx, "foo", a.templateData, id); err != nil {
		a.logger.Error(err)
		t.Fail()
	}
}

// TestUpdatesTemplateOnCreate : updates an existing template if not deleted and
// the uuid is the same.
// This test should pass.
func TestUpdateTemplateOnCreate(t *testing.T) {
	TestCreateTemplate(t)
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	id, err := uuid.Parse(ids.templateID)
	if err != nil {
		t.Fail()
	}
	if err = a.tinkDb.CreateTemplate(a.ctx, "bar", a.templateData, id); err != nil {
		a.logger.Error(err)
		t.Fail()
	}
}

// TestConcurrentCreateTemplate : inserts five records concurrently and validates
// whether or not all five records are actually committed.
func TestConcurrentCreateTemplate(t *testing.T) {
	var err error
	count := 10
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if _, err := a.tinkDb.instance.Query("delete from template;"); err != nil {
		t.Fail()
	}
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		id, err := uuid.NewRandom()
		if err != nil {
			t.Fail()
		}
		label := fmt.Sprintf("temp%v", i)
		t.Logf("PUSHING %v", i)
		if err != nil {
			t.Fail()
		}
		go func(k string, v string, a *testArgs, wg *sync.WaitGroup) {
			defer wg.Done()
			id, err := uuid.Parse(v)
			if err != nil {
				t.Fail()
			}
			if err := a.tinkDb.CreateTemplate(a.ctx, k, a.templateData, id); err != nil {
				a.logger.Error(err)
				t.Fail()
			}
		}(label, id.String(), a, &wg)
	}
	wg.Wait()
	var resultRole int
	if err = a.tinkDb.instance.QueryRow("select count(*) as count from template;").Scan(&resultRole); err != nil {
		t.Fail()
	}
	if resultRole != count {
		t.Error(fmt.Errorf("Tempaltes PUSHED: %v. \n MACS RETURNED: %v", count, resultRole))
	}
}

// TestGetTemplate : validates retrieving a workflow template.
func TestGetTemplate(t *testing.T) {
	TestCreateTemplate(t)
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	filter := make(map[string]string)
	filter["name"] = "foo"
	if _, _, _, err = a.tinkDb.GetTemplate(a.ctx, filter); err != nil {
		a.logger.Error(err)
		t.Fail()
	}
}

// TestDeleteTemplate : validates the deletion of a template.
func TestDeleteTemplate(t *testing.T) {
	setupInsertIntoDB(t)
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	if err = a.tinkDb.DeleteTemplate(a.ctx, ids.templateID); err != nil {
		a.logger.Error(err)
		t.Fail()
	}
}

// TestUpdateTemplate : validates updating a template.
func TestUpdateTemplate(t *testing.T) {
	TestCreateTemplate(t)
	var err error
	a := newTestParams(t, ids)
	defer a.tinkDb.instance.Close()

	id, err := uuid.Parse(ids.templateID)
	if err != nil {
		t.Fail()
	}
	filter := make(map[string]string)
	filter["name"] = "foo"
	if a.tinkDb.instance, err = sql.Open("postgres", a.psqlInfo); err != nil {
		t.Fail()
	}
	if err = a.tinkDb.UpdateTemplate(a.ctx, "foo", a.templateDataUpdated, id); err != nil {
		a.logger.Error(err)
		t.Fail()
	}
	_, _, received, err := a.tinkDb.GetTemplate(a.ctx, filter)
	if err != nil {
		a.logger.Error(err)
		t.Fail()
	}
	if received != a.templateDataUpdated {
		err = fmt.Errorf("EXPECTED: %s\nRECEIVED: %s", a.templateDataUpdated, received)
		t.Error(err)
	}
}
